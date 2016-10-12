//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 03/10/2016
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

// IshtmCmd base command to perform ishtm commands, lets user
// upload, get, ping and delete data to server instance.
var IshtmCmd = &cobra.Command{
	Use:       "ishtm",
	Short:     "\"If something happens to me\" service",
	Long:      "\"If something happens to me\" let you upload a \"will\" to the server and plan it's delivery... just in case.",
	ValidArgs: []string{"create", "get", "ping", "delete"},
}

func init() {
	// API references
	setArgument(IshtmCmd, "ishtmeaddress")
	setArgument(IshtmCmd, "ishtmport")

	viper.BindPFlag(am["ishtmeaddress"].name, IshtmCmd.PersistentFlags().Lookup(am["ishtmeaddress"].name))
	viper.BindPFlag(am["ishtmport"].name, IshtmCmd.PersistentFlags().Lookup(am["ishtmport"].name))

	RootCmd.AddCommand(IshtmCmd)

	// files parameters
	IshtmCmd.RunE = ishtm
}

// ishtm command expose an empty base command that should
// be called with a command option.
func ishtm(cmd *cobra.Command, args []string) error {
	return nil
}
