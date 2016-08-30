//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package main

// Go standard libraries
import (
	"fmt"
	"net/http"
)

// 3n4 libraries
import (
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/logo"
	ver "github.com/nexocrew/3nigm4/lib/version"
)

// Third party libraries
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
	Use:   "3n4chatserver",
	Short: "3nigm4 chat services",
	Long:  "Command line 3nigm4 chat server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo(nil)
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("3nigm4_CS", true, true)
	serviceAddress := fmt.Sprintf("%s:%d", "localhost", arguments.port) // arguments.address

	info := make(map[string]string)
	info["bind"] = serviceAddress

	// print logo
	printLogo(info)

	if arguments.sslcrt != "" && arguments.sslpvk != "" {
		log.MessageLog("Starting listening with TLS on address %s.\n", serviceAddress)
		// set up SSL/TLS
		fmt.Println("listening")
		err := http.ListenAndServeTLS(serviceAddress, arguments.sslcrt, arguments.sslpvk, nil)
		fmt.Println("listen returned", err.Error())
		if err != nil {

			// fmt.Errorf("https unable to listen and serve on address: %s cause error: %s", serviceAddress, err.Error())
		}
	}
}

func printLogo(i map[string]string) {
	fmt.Printf("%s", logo.Logo("3n4ChatService", ver.V().VersionString(), i))
}
