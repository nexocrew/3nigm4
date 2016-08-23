//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package main

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

// RootCmd is the base command used
// by cobra in the chatservice exec.
var RootCmd = &cobra.Command{
	Use:   "3nigm4chat",
	Short: "3nigm4 chat services",
	Long:  "Command line 3nigm4 chat server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo()
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("3nigm4_CS", true, true)
	serviceAddress := fmt.Sprintf("%s:%d", "localhost", arguments.port) // arguments.address

	if arguments.sslcrt != "" && arguments.sslpvk != "" {
		log.MessageLog("Starting listening with TLS on address %s.\n", serviceAddress)
		// set up SSL/TLS
		err = http.ListenAndServeTLS(serviceAddress, arguments.sslcrt, arguments.sslpvk, nil)
		if err != nil {

			// fmt.Errorf("https unable to listen and serve on address: %s cause error: %s", serviceAddress, err.Error())
		}
	}
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("Command line client app", ver.V().VersionString(), nil))
}
