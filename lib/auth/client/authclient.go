//
// 3nigm4 storageservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package authclient

// Std golang packages
import (
	"fmt"
	"net/rpc"
)

// 3n4 libraries
import (
	t "github.com/nexocrew/3nigm4/lib/auth/types"
)

// AuthClient is the interface used to interact
// with authentication services.
type AuthClient interface {
	Login(string, string) ([]byte, error)                       // manage user's login;
	Logout([]byte) ([]byte, error)                              // manage user's logout;
	AuthoriseAndGetInfo([]byte) (*t.UserInfoResponseArg, error) // returns authenticated user infos or an error;
	Close() error                                               // closes eventual connections.
}

// AuthRpc implements the RPC default client for
// the 3nigm4 auth service.
type AuthRpc struct {
	client *rpc.Client
}

// NewAuthRpc creates a new instance of the RPC
// client used to interact with the auth service.
func NewAuthRpc(addr string, port int) (*AuthRpc, error) {
	address := fmt.Sprintf("%s:%d", addr, port)
	rawClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}
	return &AuthRpc{
		client: rawClient,
	}, nil
}

// Login grant access to users, over RPC, using username and password.
func (a *AuthRpc) Login(username string, password string) ([]byte, error) {
	// perform login on RPC service
	var loginResponse t.LoginResponseArg
	err := a.client.Call("Login.Login", &t.LoginRequestArg{
		Username: username,
		Password: password,
	}, &loginResponse)
	if err != nil {
		return nil, err
	}
	return loginResponse.Token, nil
}

// Logout remove actual active sessions over RPC.
func (a *AuthRpc) Logout(token []byte) ([]byte, error) {
	var logoutResponse t.LogoutResponseArg
	err := a.client.Call("Login.Logout", &t.LogoutRequestArg{
		Token: token,
	}, &logoutResponse)
	if err != nil {
		return nil, err
	}
	return logoutResponse.Invalidated, nil
}

// AuthoriseAndGetInfo if the token is valid returns info about
// the associated user over RPC service.
func (a *AuthRpc) AuthoriseAndGetInfo(token []byte) (*t.UserInfoResponseArg, error) {
	// verify token and retrieve user infos
	var authResponse t.UserInfoResponseArg
	err := a.client.Call("SessionAuth.UserInfo", &t.AuthenticateRequestArg{
		Token: token,
	}, &authResponse)
	if err != nil {
		return nil, err
	}
	return &authResponse, nil
}

// Close closes RPC connection.
func (a *AuthRpc) Close() error {
	return a.client.Close()
}