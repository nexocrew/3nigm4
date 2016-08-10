//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Internal libs
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	ct "github.com/nexocrew/3nigm4/lib/commontypes"
)

// Third party
import (
	"github.com/gorilla/mux"
)

// riseError rises an error returning a standard error
// message.
func riseError(status int, msg string, w http.ResponseWriter, ipa string) {
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(
		ct.StandardResponse{
			ct.NakResponse,
			msg,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.ErrorLog("Request processing error: %s from IP %s.\n", msg, ipa)
	}
}

// authoriseGettingUserInfos authorises the provided token
// and return user associated data. If returns a nil value
// it means something went wrong.
func authoriseGettingUserInfos(authToken string) (*auth.UserInfoResponseArg, error) {
	// verify token and retrieve user infos
	var authResponse auth.UserInfoResponseArg
	if authToken == "" {
		return nil, fmt.Errorf("authorisation token is nil")
	}

	token, err := hex.DecodeString(authToken)
	if err != nil {
		return nil, fmt.Errorf("authorisation token is malformed (%s)", err.Error())
	}
	err = rpcClient.Call("SessionAuth.UserInfo", &auth.AuthenticateRequestArg{
		Token: token,
	}, &authResponse)
	if err != nil {
		return nil, err
	}
	return &authResponse, nil
}

// postChunk upload a data chunk to the S3 backend service
// after authorising the user. it interacts in sync with multiple
// services in order to obtain user authentication, s3 backend and
// database logging functionalities.
func postChunk(w http.ResponseWriter, r *http.Request) {
	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err == nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// get message BODY
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()
	// parse json body
	var requestBody ct.SechunkPostRequest
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
	if requestBody.Data == nil ||
		len(requestBody.Data) == 0 {
		riseError(http.StatusBadRequest,
			"data in request body is nil", w,
			r.RemoteAddr)
		return
	}

	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()

	// insert file log in the database
	checksum := sha256.Sum256(requestBody.Data)
	fl := &FileLog{
		Id:         requestBody.Id,
		Size:       len(requestBody.Data),
		Bucket:     arguments.s3Bucket,
		Creation:   time.Now(),
		TimeToLive: requestBody.TimeToLive,
		CheckSum: ct.CheckSum{
			Hash: checksum[:],
			Type: "SHA256",
		},
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
		Acl: Acl{
			Permission:   requestBody.Permission,
			SharingUsers: requestBody.SharingUsers,
		},
	}
	err = dbSession.SetFileLog(fl)
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        fl.Id,
		Complete:  false,
		TimeStamp: fl.Creation,
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// upload data to the S3 backend
	var expireTime *time.Time = nil
	if fl.TimeToLive != 0 {
		ttl := fl.Creation.Add(fl.TimeToLive)
		expireTime = &ttl
	}
	s3backend.Upload(fl.Bucket, fl.Id, requestBody.Data, expireTime)

	// return upload response message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.SechunkPostResponse{
			Id: fl.Id,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Upload request %s accepted, waiting for upload verification", fl.Id)
	}
}

func getChunk(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
}

func deleteChunk(w http.ResponseWriter, r *http.Request) {

}

// getVerifyTx responds to a request of info regarding a
// previously produced async request (upload, download of
// delete). It returns available infos based on the originally
// required operation (for example Data field is only available
// on download requests). After completing the query flow it
// removes the async tx record.
func getVerifyTx(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get tx status
	at, err := dbSession.GetAsyncTx(id)
	if err != nil {
		riseError(http.StatusNotFound,
			fmt.Sprintf("unable to find required request, verification must be done at max %d min from order request", MaxAsyncTxExistance),
			w, r.RemoteAddr)
		return
	}
	// clean db tx
	if at.Complete == true {
		err = dbSession.RemoveAsyncTx(id)
		if arguments.verbose &&
			err != nil {
			log.WarningLog("Unable to remove async tx from database: %s.\n", err.Error())
		}
	}

	// return verify message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.SechunkTxVerify{
			Complete: at.Complete,
			Error:    at.Error.Error(),
			Data:     at.Data,
			CheckSum: at.CheckSum,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Verify request %s correcly replyied.\n", id)
	}
}

// Ping function to verify if the service is on
// or not.
func getPing(w http.ResponseWriter, r *http.Request) {
	/* return value */
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&ct.StandardResponse{
		Status: ct.AckResponse,
	})
}
