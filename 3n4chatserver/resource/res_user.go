//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package resource

// Go standard libraries
import (
	"net/http"
)

// 3n4 libraries
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

type Auth struct {
	AuthClient *auth.AuthRPC
}

// UserCredential is the JSON
// body posted by client via /login API
type UserCredential struct {
	Username string `json:"usr"`
	Password string `json:"pwd"`
}

// LoginResponse contains the
// user session token
type LoginResponse struct {
	TokenBase64 string `json:"stk"`
}

// Post Login API
// this API expects in the request headers
// a valid username and hashed password
func (a *Auth) Get(r *http.Request) (int, h.Resource) {
	// check for valid headers
	username := r.Header.Get(usernameHttpHeader)
	password := r.Header.Get(passwordHttpHeader)

	// authenticate user via RPC service
	token, err := a.AuthClient.Login(username, password)
	if err != nil {
		return http.StatusBadRequest, nil
	}

	// generate sessionToken and send reply
	return http.StatusOK, token
}

func (a *Auth) Delete(r *http.Request) (int, h.Resource) {
	return http.StatusOK, nil
}
