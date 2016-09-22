//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"os"
)

// Internal pkgs
import (
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/logo"
	ver "github.com/nexocrew/3nigm4/lib/version"
)

// Third party pkgs
import (
	"github.com/spf13/cobra"
)

// Logger global instance
var log *logger.LogFacility

// Cobra parsed arguments
var arguments args

// RootCmd is the base command used by cobra in the ishtmservice
// exec.
var RootCmd = &cobra.Command{
	Use:   "ishtmservice",
	Short: "If Something Happens To Me REST API backend",
	Long:  "Server that interface via REST API the access for the If Something Happens To Me service.",
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

// Execute parsing and execute selected
// command.
func Execute() error {
	// execute actual command
	_, err := RootCmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("Ishtm REST API backend server", ver.V().VersionString(), nil))
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("ishtmservice", true, true)

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
