//
// 3nigm4 3n4cli package
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
	Example: "3n4cli logout",
}

// logout closes the authenticated session created by the
// login command. The global var token is set to nil after
// requiring the server to dismiss the session.
func logout(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)

	if pss.Token == "" {
		return fmt.Errorf("no token set, unable to logout user")
	}

	// create http request
	client := &http.Client{}
	// get address and port
	authAddress := viper.GetString(viperLabel(cmd, "authaddress"))
	authPort := viper.GetInt(viperLabel(cmd, "authport"))
	// prepare post request
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s:%d%s", authAddress, authPort, authorisationPath),
		nil)
	if err != nil {
		return err
	}
	token := pss.Token
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
	pss.invalidateSessionToken()

	// if verbose printf token
	if viper.GetBool(viperLabel(RootCmd, "verbose")) {
		log.VerboseLog("Invalidated token: %s.\n", token)
	}

	return nil
}
