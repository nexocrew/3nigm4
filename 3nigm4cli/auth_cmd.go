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
	"github.com/spf13/viper"
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
	// API references
	LoginCmd.PersistentFlags().StringVarP(&arguments.authService.Address, "authaddrs", "", "https://www.nexo.cloud", "the authentication service address")
	LoginCmd.PersistentFlags().IntVarP(&arguments.authService.Port, "authport", "", 443, "the authentication service port")
	LoginCmd.PersistentFlags().StringVarP(&arguments.username, "username", "u", "", "the user requesting to login")

	viper.BindPFlag("AuthServiceAddress", LoginCmd.PersistentFlags().Lookup("authaddrs"))
	viper.BindPFlag("AuthServicePort", LoginCmd.PersistentFlags().Lookup("authport"))
	viper.BindPFlag("Username", LoginCmd.PersistentFlags().Lookup("username"))

	viper.SetDefault("AuthServiceAddress", "https://www.nexo.cloud")
	viper.SetDefault("AuthServicePort", 443)

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
