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
	ct "github.com/nexocrew/3nigm4/lib/commons"
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
	if authToken == "" {
		return nil, fmt.Errorf("authorisation token is nil")
	}

	token, err := hex.DecodeString(authToken)
	if err != nil {
		return nil, fmt.Errorf("authorisation token is malformed (%s)", err.Error())
	}
	authResponse, err := authClient.AuthoriseAndGetInfo(token)
	if err != nil {
		return nil, err
	}
	return authResponse, nil
}

// login is used to provide login functionality based on the
// auth service.
func login(w http.ResponseWriter, r *http.Request) {
	// get message BODY
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()
	// parse json body
	var requestBody ct.LoginRequest
	err := json.Unmarshal(body, &requestBody)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
	if requestBody.Username == "" ||
		requestBody.Password == "" {
		riseError(http.StatusBadRequest,
			"username or password in request body are nil", w,
			r.RemoteAddr)
		return
	}

	// perform login on auth service
	token, err := authClient.Login(requestBody.Username, requestBody.Password)
	if err != nil {
		riseError(http.StatusUnauthorized,
			"unable to login with provided credentials", w,
			r.RemoteAddr)
		return
	}
	// return the session token
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.LoginResponse{
			Token: hex.EncodeToString(token),
		})
	if err != nil {
		panic(err)
	}
}

// logout implements, redirecting to auth service, the
// user's session invalidation.
func logout(w http.ResponseWriter, r *http.Request) {
	authToken := r.Header.Get(ct.SecurityTokenKey)
	if authToken == "" {
		riseError(http.StatusBadRequest,
			"auth token is nil", w,
			r.RemoteAddr)
		return
	}

	rawToken, err := hex.DecodeString(authToken)
	if err != nil {
		riseError(http.StatusBadRequest,
			"auth token is malformed", w,
			r.RemoteAddr)
		return
	}
	invalidated, err := authClient.Logout(rawToken)
	if err != nil {
		riseError(http.StatusUnauthorized,
			"unable to logout with provided credentials", w,
			r.RemoteAddr)
		return
	}
	// return the session token
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.LogoutResponse{
			Invalidated: hex.EncodeToString(invalidated),
		})
	if err != nil {
		panic(err)
	}
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

	// get time stamp
	now := time.Now()
	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// insert file log in the database
	checksum := sha256.Sum256(requestBody.Data)
	fl := &FileLog{
		Id:         requestBody.ID,
		Size:       len(requestBody.Data),
		Bucket:     arguments.s3Bucket,
		Creation:   now,
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

	// generate tx id
	txId := generateTranscationId(fl.Id, userInfo.Username, &now)

	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        txId,
		Complete:  false,
		TimeStamp: fl.Creation,
		Ownership: fl.Ownership,
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
	s3backend.Upload(fl.Bucket, fl.Id, txId, requestBody.Data, expireTime)

	// return upload response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.SechunkAsyncResponse{
			ID: txId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Upload request %s accepted, waiting for upload verification", txId)
	}
}

// checkAclPermission verify all possible acl scenarios and check if the
// requiring user has required permissions to access the file. If user can
// download it it'll return true otherwise false.
func checkAclPermission(userInfo *auth.UserInfoResponseArg, fileLog *FileLog) bool {
	// check access credentials
	switch fileLog.Acl.Permission {
	case Private:
		if fileLog.Ownership.Username == userInfo.Username {
			return true
		}
	case Public:
		return true
	case Shared:
		for _, permitted := range fileLog.Acl.SharingUsers {
			if permitted == userInfo.Username {
				return true
			}
		}
	}
	return false
}

// getChunk implements the first step of a file download request it is exposed
// via a REST GET method and returns a txId usable with the verify API call to
// retrieve the actual downloaded data (from S3 storage). The user must be
// correctly authenticated to be able to access the requested resource.
func getChunk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["resourceid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err == nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get resources info
	fileLog, err := dbSession.GetFileLog(id)
	if err != nil {
		riseError(http.StatusNotFound,
			fmt.Sprintf("requested file not found"), w,
			r.RemoteAddr)
		return
	}

	// check permission
	granted := checkAclPermission(userInfo, fileLog)
	if !granted {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("you are not authorised to access this resource"), w,
			r.RemoteAddr)
		return
	}

	now := time.Now()
	// generate tx id
	txId := generateTranscationId(id, userInfo.Username, &now)
	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        txId,
		Complete:  false,
		TimeStamp: now,
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// require S3 download
	s3backend.Download(arguments.s3Bucket, id, txId)

	// return download response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.SechunkAsyncResponse{
			ID: txId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Download request %s accepted, waiting for download verification", txId)
	}
}

// deleteChunk remove a file from the S3 storage: only the original file
// owner (who uploaded it) can remove a file from there.
func deleteChunk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["resourceid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err == nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get resources info
	fileLog, err := dbSession.GetFileLog(id)
	if err != nil {
		riseError(http.StatusNotFound,
			fmt.Sprintf("requested file not found"), w,
			r.RemoteAddr)
		return
	}

	// check permission
	var granted bool
	// strict acl verification: only the file owner is able
	// to delete it.
	if fileLog.Ownership.Username == userInfo.Username {
		granted = true
	}
	if !granted {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("you are not authorised to delete this resource"), w,
			r.RemoteAddr)
		return
	}

	now := time.Now()
	// generate tx id
	txId := generateTranscationId(id, userInfo.Username, &now)
	// add async tx record
	err = dbSession.SetAsyncTx(&AsyncTx{
		Id:        txId,
		Complete:  false,
		TimeStamp: now,
		Ownership: Owner{
			Username:  userInfo.Username,
			OriginIp:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
	})
	if err != nil {
		riseError(http.StatusInternalServerError,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// require S3 download
	s3backend.Delete(arguments.s3Bucket, id, txId)

	// return download response message
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(
		&ct.SechunkAsyncResponse{
			ID: txId,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Delete request %s accepted, waiting for delete verification", txId)
	}
}

// getQueue responds to a request of info regarding a
// previously produced async request (upload, download of
// delete). It returns available infos based on the originally
// required operation (for example Data field is only available
// on download requests). After completing the query flow it
// removes the async tx record.
func getQueue(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["requestid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err == nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// retain db
	dbSession := db.Copy()
	defer dbSession.Close()
	// get tx status
	at, err := dbSession.GetAsyncTx(id)
	if err != nil {
		riseError(http.StatusGone,
			fmt.Sprintf("unable to find required request, verification must be done at max %d min from order request", MaxAsyncTxExistance),
			w, r.RemoteAddr)
		return
	}

	// check ownership
	if at.Ownership.Username != userInfo.Username {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("user is not authorised to verify %s request id", id), w,
			r.RemoteAddr)
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
