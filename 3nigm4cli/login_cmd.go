//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import ()

// Internal dependencies
import ()

// Third party libs
import (
	"github.com/spf13/cobra"
)

// Global var used to store the login output
// token, this should be used to check if the user is
// logged in or not.
var token string

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

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func login(cmd *cobra.Command, args []string) error {
	// load config file
	err := manageConfigFile()
	if err != nil {
		return err
	}

	return nil
}
