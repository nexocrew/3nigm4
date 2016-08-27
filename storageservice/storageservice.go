//
// 3nigm4 storageservice package
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
	Use:   "storageservice",
	Short: "Storage REST API backend",
	Long:  "Server that interface via REST API the access for file storage on S3 backend.",
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
	RootCmd.AddCommand(ServeCmd)
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
	fmt.Printf("%s", logo.Logo("Storage REST API backend server", ver.V().VersionString(), nil))
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("storageservice", true, true)

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
