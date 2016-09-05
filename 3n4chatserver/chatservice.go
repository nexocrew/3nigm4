//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package main

// Go standard libraries
import (
	"fmt"
	"os"
)

// 3n4 libraries
import (
	"github.com/nexocrew/3nigm4/lib/auth"
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
var a args

// AuthRPC client
var authClient *auth.AuthRPC

// init a cobra command
var rootCmd = &cobra.Command{
	Use:   "3n4chatserver",
	Short: "3nigm4 chat",
	Long:  "Command line 3nigm4 chat server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo(nil)
		return fmt.Errorf("undefined command, select a valid one\n")
	},
}

func init() {
	// start up logging facility
	log = logger.NewLogFacility("3n4CHAT", true, true)

	// set flags
	rootCmd.PersistentFlags().BoolVarP(&a.verbose, "verbose", "v", false, "enable verbose mode")
	// database references
	rootCmd.PersistentFlags().StringVarP(&a.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	rootCmd.PersistentFlags().StringVarP(&a.dbUsername, "dbuser", "u", "", "the database user name")
	rootCmd.PersistentFlags().StringVarP(&a.dbPassword, "dbpwd", "w", "", "the database password")
	rootCmd.PersistentFlags().StringVarP(&a.dbAuth, "dbauth", "", "admin", "the database auth db")
	// auth RPC service
	rootCmd.PersistentFlags().StringVarP(&a.authServiceAddress, "authaddr", "A", "", "the authorisation RPC service address")
	rootCmd.PersistentFlags().IntVarP(&a.authServicePort, "authport", "P", 7931, "the authorisation RPC service port")
}

func main() {
	// print logo
	printLogo(nil)

	// execute command
	rootCmd.AddCommand(serveCmd)
	_, err := rootCmd.ExecuteC()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func printLogo(i map[string]string) {
	fmt.Printf("%s", logo.Logo("3n4ChatService", ver.V().VersionString(), i))
}
