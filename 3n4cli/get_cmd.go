//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 04/10/2016
//

package main

// Golang std libs
import ()

// Internal dependencies
import (
	_ "github.com/nexocrew/3nigm4/lib/commons"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmd get a will record (totally or partially)
// from a will ID,
var GetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get and download a \"will\" activity",
	Long:    "Get infos and download a \"will\" activity record.",
	Example: "3n4cli ishtm get",
	PreRun:  verbosePreRunInfos,
}

func init() {
	// i/o paths
	setArgument(GetCmd, "output")

	viper.BindPFlag(am["output"].name, GetCmd.PersistentFlags().Lookup(am["output"].name))

	IshtmCmd.AddCommand(GetCmd)

	// files parameters
	GetCmd.RunE = get
}

func get(cmd *cobra.Command, args []string) error {
	return nil
}
