//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"bytes"
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

// postJob creates a new async job request passing in the body
// the requested command to be executed with all required
// arguments.
func postJob(w http.ResponseWriter, r *http.Request) {
	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err != nil {
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
	var job ct.JobPostRequest
	err = json.Unmarshal(body, &job)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
	// check for arguments
	if job.Arguments == nil ||
		job.Arguments.ResourceID == "" {
		riseError(http.StatusBadRequest,
			"unable to process requests with nil arguments", w,
			r.RemoteAddr)
		return
	}
	// TODO: eventually add a regex to validate the
	// passed resource id.

	// select the right command implementation
	switch job.Command {
	case "UPLOAD":
		createStorageResource(w, r, &job, userInfo)
	case "DOWNLOAD":
		retrieveStorageResource(w, r, &job, userInfo)
	case "DELETE":
		deleteStorageResource(w, r, &job, userInfo)
	default:
		riseError(http.StatusBadRequest,
			"unknown command", w,
			r.RemoteAddr)
		return
	}
}

// getJob responds to a request of info regarding a
// previously produced async request (upload, download of
// delete). It returns available infos based on the originally
// required operation (for example Data field is only available
// on download requests). After completing the query flow it
// removes the async tx record.
func getJob(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["jobid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err != nil {
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
			fmt.Sprintf("unable to find %s job, verification must be done at max %d min from request creation", id, MaxAsyncTxExistance/time.Minute),
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

	// in case the processing is not yet completed
	if at.Complete == false {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	var errstr string
	if at.Error != nil {
		errstr = at.Error.Error()
	}
	// return verify message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.JobGetRequest{
			Complete: at.Complete,
			Error:    errstr,
			Data:     at.Data,
			CheckSum: at.CheckSum,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Verify request %s correcly replyied.\n", id)
	}

	// remove obsolete job db document
	err = dbSession.RemoveAsyncTx(id)
	if arguments.verbose &&
		err != nil {
		log.WarningLog("Unable to remove async tx from database: %s.\n", err.Error())
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
