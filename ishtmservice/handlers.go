//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

// Golang std pkgs
import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// Internal pkgs
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm"
)

// Third party pkgs
import (
	_ "github.com/gorilla/mux"
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

func postWill(w http.ResponseWriter, r *http.Request) {
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
	var willRequest ct.WillPostRequest
	err = json.Unmarshal(body, &willRequest)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
	// check for arguments
	if willRequest.Reference == nil ||
		len(willRequest.Reference) == 0 ||
		willRequest.Recipients == nil ||
		len(willRequest.Recipients) == 0 {
		riseError(http.StatusBadRequest,
			"wrong request body values", w,
			r.RemoteAddr)
		return
	}

	// create will
	will := ishtm.NewWill()

}

func getWill(w http.ResponseWriter, r *http.Request) {
	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err != nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
}

func patchWill(w http.ResponseWriter, r *http.Request) {
	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err != nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	// authorise and get user's info
	// extract token from headers
	userInfo, err := authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
	if err != nil {
		riseError(http.StatusUnauthorized,
			err.Error(), w,
			r.RemoteAddr)
		return
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
