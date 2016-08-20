//
// 3nigm4 3nigm4cli package
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
	"io/ioutil"
	"net/http"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global var used to store the login output
// token, this should be used to check if the user is
// logged in or not.
var token string

const (
	authorisationPath = "/v1/authsession"
)

// LoginCmd let's the user connect to auth server
// and login.
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login a registered user and manage session",
	Long:    "Interact with authentication server to login at the application startup.",
	Example: "3nigm4cli login -u username",
}

func init() {
	setArgument(LoginCmd, "authaddress", &arguments.authService.Address)
	setArgument(LoginCmd, "authport", &arguments.authService.Port)
	setArgument(LoginCmd, "username", &arguments.username)

	// files parameters
	LoginCmd.RunE = login
}

// hexComposedPassword compose a string and checksum
// it to avoid pass over the internet a plain
// text password.
func hexComposedPassword(username string, pwd []byte) string {
	var composed []byte
	header := fmt.Sprintf("3nigm4.%s.", username)
	composed = append(composed, []byte(header)...)
	composed = append(composed, pwd...)
	checksum := sha256.Sum256(composed)
	hexPwd := hex.EncodeToString(checksum[:])
	return hexPwd
}

// login command let the user authenticate on all available
// 3nigm4 services, this function will be called before any
// other to be able to proceed with a valid auth token.
func login(cmd *cobra.Command, args []string) error {
	// load config file
	err := manageConfigFile()
	if err != nil {
		return err
	}

	username := viper.GetString(am["username"].name)
	// get user password
	pwd, err := gopass.GetPasswd()
	if err != nil {
		return err
	}

	// prepare request
	lr := &ct.LoginRequest{
		Username: username,
		Password: hexComposedPassword(username, pwd),
	}
	body, err := json.Marshal(lr)
	if err != nil {
		return err
	}

	// create http request
	client := &http.Client{}
	// get address and port
	authAddress := viper.GetString(am["authaddress"].name)
	authPort := viper.GetInt(am["authport"].name)
	// prepare post request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s:%d%s", authAddress, authPort, authorisationPath),
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// get token
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// check for errors
	err = checkRequestStatus(resp.StatusCode, http.StatusOK, respBody)
	if err != nil {
		return err
	}

	var loginResponse ct.LoginResponse
	err = json.Unmarshal(respBody, &loginResponse)
	if err != nil {
		return err
	}
	if loginResponse.Token == "" {
		return fmt.Errorf("returned token is nil, unable to proceed")
	}
	// set global token with returned one
	token = loginResponse.Token

	return nil
}

// logout closes the authenticated session created by the
// login command. The global var token is set to nil after
// requiring the server to dismiss the session.
func logout() error {
	if token == "" {
		return fmt.Errorf("no token set, unable to logout user")
	}

	// create http request
	client := &http.Client{}
	// get address and port
	authAddress := viper.GetString(am["authaddress"].name)
	authPort := viper.GetInt(am["authport"].name)
	// prepare post request
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s:%d%s", authAddress, authPort, authorisationPath),
		nil)
	if err != nil {
		return err
	}
	req.Header.Set(ct.SecurityTokenKey, token)
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// get token
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// check for errors
	err = checkRequestStatus(resp.StatusCode, http.StatusOK, respBody)
	if err != nil {
		return err
	}

	var logoutResponse ct.LogoutResponse
	err = json.Unmarshal(respBody, &logoutResponse)
	if err != nil {
		return err
	}
	// set global token to nil
	token = ""

	return nil
}
