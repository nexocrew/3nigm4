//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"fmt"
	"os"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/logo"
	ver "github.com/nexocrew/3nigm4/lib/version"
)

// Third party libs
import (
	"github.com/spf13/cobra"
)

// Logger global instance
var log *logger.LogFacility

// Cobra parsed arguments
var arguments args

// RootCmd is the base command used by cobra in the storageservice
// exec.
var RootCmd = &cobra.Command{
	Use:   "3nigm4cli",
	Short: "CLI client for the 3nigm4 services",
	Long:  "Command line client to access 3nigm4 services, it generally requires a network connection to operate.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo()
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func init() {
	// global flags
	RootCmd.PersistentFlags().BoolVarP(&arguments.verbose, "verbose", "v", false, "activate logging verbosity")
	RootCmd.PersistentFlags().BoolVarP(&arguments.colored, "colored", "C", true, "activate colored logs")
}

// AddCommands adds available commands
// to the root command
func AddCommands() {
	RootCmd.AddCommand(StoreCmd)
}

// Execute parsing and execute selected
// command.
func Execute() error {
	// add commands
	AddCommands()

	// execute actual command
	_, err := RootCmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("Command line client app", ver.V().VersionString(), nil))
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("3nigm4cli", true, true)

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
