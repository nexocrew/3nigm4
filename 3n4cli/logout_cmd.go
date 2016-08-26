//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LoginCmd let's the user connect to auth server
// and login.
var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout a previously generated session",
	Long:    "Interact with authentication server to logout a previously generated session token.",
	Example: "3nigm4cli logout",
}

func init() {
	setArgument(LogoutCmd, "authaddress", &arguments.authService.Address)
	setArgument(LogoutCmd, "authport", &arguments.authService.Port)

	// files parameters
	LogoutCmd.RunE = logout
}

// logout closes the authenticated session created by the
// login command. The global var token is set to nil after
// requiring the server to dismiss the session.
func logout(cmd *cobra.Command, args []string) error {
	if pss.Token == "" {
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
	req.Header.Set(ct.SecurityTokenKey, pss.Token)
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
	pss.invalidateSessionToken()

	return nil
}
