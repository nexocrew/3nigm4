//
// 3nigm4 3n4cli package
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

const (
	authorisationPath = "/v1/authsession"
)

// LoginCmd let's the user connect to auth server
// and login.
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login a registered user and manage session",
	Long:    "Interact with authentication server to login at the application startup.",
	Example: "3n4cli login -u username",
	PreRun:  verbosePreRunInfos,
}

func init() {
	RootCmd.AddCommand(LoginCmd)

	setArgument(LoginCmd, "authaddress")
	setArgument(LoginCmd, "authport")
	setArgument(LoginCmd, "username")

	bindPFlag(LoginCmd, "username")
	bindPFlag(LoginCmd, "authaddress")
	bindPFlag(LoginCmd, "authport")

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
	username := viper.GetString(viperLabel(cmd, "username"))
	// get user password
	fmt.Printf("Insert password: ")
	pwd, err := gopass.GetPasswdMasked()
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
	authAddress := viper.GetString(viperLabel(cmd, "authaddress"))
	authPort := viper.GetInt(viperLabel(cmd, "authport"))
	// prepare post request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s:%d%s", authAddress, authPort, authorisationPath),
		bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to create the request %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform the request cause %s", err.Error())
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
	pss.Token = loginResponse.Token
	pss.refreshLastLogin()

	// if verbose printf token
	if viper.GetBool(viperLabel(RootCmd, "verbose")) {
		log.VerboseLog("Token obtained: %s.\n", pss.Token)
	}

	return nil
}
