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
	"strconv"
	"time"
)

// Internal pkgs
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm"
)

// Third party pkgs
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

var (
	additionalExpiration = 24 * time.Hour
)

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
	will, credentials, err := ishtm.NewWill(
		&ishtm.OwnerID{
			Name:  userInfo.Username,
			Email: userInfo.Email,
		},
		willRequest.Reference,
		&ishtm.Settings{
			DeliveryOffset: additionalExpiration,
			DisableOffset:  false,
			NotifyDeadline: willRequest.NotifyDeadline,
			ExtensionUnit:  willRequest.ExtensionUnit,
		},
		willRequest.Recipients,
	)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
			r.RemoteAddr)
		return
	}

	// save it!
	dbSession := db.Copy()
	defer dbSession.Close()
	err = dbSession.SetWill(will)
	if err != nil {
		riseError(http.StatusInternalServerError,
			fmt.Sprintf("unable to save to the db %s", err.Error()), w,
			r.RemoteAddr)
		return
	}

	// return response message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.WillPostResponse{
			ID:          will.ID,
			Credentials: credentials,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Will Post request on %s correcly executed.\n", will.ID)
	}
}

func getWill(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["willid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	var userInfo *auth.UserInfoResponseArg
	var authKey []byte
	var err error
	/* get query parameters */
	query := r.URL.Query()
	key, ok := query["deliverykey"]
	if !ok {
		// authorise and get user's info
		// extract token from headers
		userInfo, err = authoriseGettingUserInfos(r.Header.Get(ct.SecurityTokenKey))
		if err != nil {
			riseError(http.StatusUnauthorized,
				err.Error(), w,
				r.RemoteAddr)
			return
		}
	} else {
		if len(key) < 1 || len(key[0]) == 0 {
			riseError(http.StatusUnauthorized,
				"unable to read delivery key", w,
				r.RemoteAddr)
			return
		}
		authKey, err = hex.DecodeString(key[0])
		if err != nil {
			riseError(http.StatusUnauthorized,
				"unable to read delivery key", w,
				r.RemoteAddr)
			return
		}
	}

	dbSession := db.Copy()
	defer dbSession.Close()
	will, err := dbSession.GetWill(id)
	if err != nil {
		riseError(http.StatusBadRequest,
			fmt.Sprintf("unable to locate will record %s", id), w,
			r.RemoteAddr)
		return
	}

	var response ct.WillGetResponse
	if authKey != nil {
		// check for delivery key and manage get for
		// recipients.
		if bytes.Compare(will.DeliveryKey, authKey) != 0 {
			riseError(http.StatusUnauthorized,
				"unable to authosize delivery key", w,
				r.RemoteAddr)
			return
		}
	} else if userInfo != nil {
		// check for logged user
		if userInfo.Username != will.Owner.Name {
			riseError(http.StatusUnauthorized,
				"unable to authosize user to access", w,
				r.RemoteAddr)
			return
		}
		response.Recipients = will.Recipients
		response.TimeToDelivery = will.TimeToDelivery
		response.ExtensionUnit = will.Settings.ExtensionUnit
		response.NotifyDeadline = will.Settings.NotifyDeadline
		response.Disabled = will.Disabled
		if will.Settings.DisableOffset != true {
			response.DeliveryOffset = will.Settings.DeliveryOffset
		}
	} else {
		// no other way of authenticate
		riseError(http.StatusUnauthorized,
			"unable to authosize access", w,
			r.RemoteAddr)
		return
	}

	response.ID = will.ID
	response.Creation = will.Creation
	response.LastModified = will.LastModified
	response.ReferenceFile = will.ReferenceFile
	response.LastPing = will.LastPing

	// return response message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&response)
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Will Get request on %s correcly executed.\n", id)
	}
}

// patchWill for now it only let the user update the count down
// value not modifying the values.
// Notice: this function could, potentially, include other
// modification to the original doc, having a critical behaviour is
// much more preferrable having the simpler create/delete pattern
// possible.
func patchWill(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["willid"]
	if !ok || id == "" {
		riseError(http.StatusBadRequest,
			"unable to proceed with nil id", w,
			r.RemoteAddr)
		return
	}

	// get message BODY
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()
	// parse json body
	var willRequest ct.WillPatchRequest
	err := json.Unmarshal(body, &willRequest)
	if err != nil {
		riseError(http.StatusBadRequest,
			err.Error(), w,
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

	dbSession := db.Copy()
	defer dbSession.Close()
	will, err := dbSession.GetWill(id)
	if err != nil {
		riseError(http.StatusBadRequest,
			fmt.Sprintf("unable to locate will record %s", id), w,
			r.RemoteAddr)
		return
	}

	if will.Owner.Name != userInfo.Username {
		riseError(http.StatusUnauthorized,
			"unable to authosize user to access", w,
			r.RemoteAddr)
		return
	}

	// validate OTP
	err = will.VerifyOtp(
		willRequest.Index,
		willRequest.Otp,
		willRequest.SecondaryKey,
	)
	if err != nil {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("unable to authorize the request %s", err.Error()), w,
			r.RemoteAddr)
		return
	}

	err = will.Refresh()
	if err != nil {
		riseError(http.StatusInternalServerError,
			fmt.Sprintf("unable proceed refreshing will record %s", err.Error()), w,
			r.RemoteAddr)
		return
	}

	// return response message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.StandardResponse{
			Status: ct.AckResponse,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Will Patch refresh request on %s correcly executed.\n", id)
	}
}

func deleteWill(w http.ResponseWriter, r *http.Request) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["willid"]
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

	/* get query parameters */
	query := r.URL.Query()
	otpValue, otpOK := query["otp"]
	secondaryValue, secondaryOK := query["secondarykey"]
	indexValue, indexOK := query["index"]
	if !otpOK && !secondaryOK {
		riseError(http.StatusUnauthorized,
			"no authentication parameters", w,
			r.RemoteAddr)
		return
	}
	var idx int
	if indexOK &&
		len(indexValue) > 0 {
		idx, err = strconv.Atoi(indexValue[0])
		if err != nil {
			riseError(http.StatusBadRequest,
				fmt.Sprintf("invalid index %s", err.Error()), w,
				r.RemoteAddr)
			return
		}
	}
	var otp, secondaryk string
	if otpValue != nil &&
		len(otpValue) > 0 {
		otp = otpValue[0]
	}
	if secondaryValue != nil &&
		len(secondaryValue) > 0 {
		secondaryk = secondaryValue[0]
	}

	dbSession := db.Copy()
	defer dbSession.Close()
	will, err := dbSession.GetWill(id)
	if err != nil {
		riseError(http.StatusBadRequest,
			fmt.Sprintf("unable to locate will record %s", id), w,
			r.RemoteAddr)
		return
	}

	if will.Owner.Name != userInfo.Username {
		riseError(http.StatusUnauthorized,
			"unable to authosize user to access", w,
			r.RemoteAddr)
		return
	}

	// validate OTP
	err = will.VerifyOtp(
		idx,
		otp,
		secondaryk,
	)
	if err != nil {
		riseError(http.StatusUnauthorized,
			fmt.Sprintf("unable to authorize the request %s", err.Error()), w,
			r.RemoteAddr)
		return
	}

	err = dbSession.RemoveWill(id)
	if err != nil {
		riseError(http.StatusInternalServerError,
			fmt.Sprintf("unable to delete will record %s", err.Error()), w,
			r.RemoteAddr)
		return
	}

	// return response message
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		&ct.StandardResponse{
			Status: ct.AckResponse,
		})
	if err != nil {
		panic(err)
	}
	if arguments.verbose {
		log.VerboseLog("Will Delete request on %s correcly executed.\n", id)
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
