//
// 3nigm4 crypto package
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

// Root command
var AuthCmd = &cobra.Command{
	Use:   "authserver",
	Short: "Authentication server",
	Long:  "Authentication server is employed by 3nigm4 server applications to verify user's identity.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo()
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func init() {
	// global flags
	AuthCmd.PersistentFlags().BoolVarP(&arguments.verbose, "verbose", "v", false, "activate logging verbosity")
	AuthCmd.PersistentFlags().BoolVarP(&arguments.colored, "colored", "C", true, "activate colored logs")
}

// AddCommands adds available commands
// to the root command
func AddCommands() {
	AuthCmd.AddCommand(ServeCmd)
}

// Execute parsing and execute selected
// command.
func Execute() error {
	// add commands
	AddCommands()

	// execute actual command
	_, err := AuthCmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("AuthServe: RPC based authentication server", ver.V().VersionString(), nil))
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("authserver", true, true)

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
