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
	"os"
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
var a args

func init() {
	// start up logging facility
	log = logger.NewLogFacility("3n4CHAT", true, true)
	// init a cobra command
	rootCmd := &cobra.Command{
		Use:   "3n4chatserver",
		Short: "a",
		Long:  "Command line 3nigm4 chat server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			printLogo(nil)
			return fmt.Errorf("undefined command, select a valid one\n")
		},
	}
	// set flags
	rootCmd.PersistentFlags().BoolVarP(&a.verbose, "verbose", "v", false, "enable verbose mode")
	// database references
	rootCmd.PersistentFlags().StringVarP(&a.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	rootCmd.PersistentFlags().StringVarP(&a.dbUsername, "dbuser", "u", "", "the database user name")
	rootCmd.PersistentFlags().StringVarP(&a.dbPassword, "dbpwd", "w", "", "the database password")
	rootCmd.PersistentFlags().StringVarP(&a.dbAuth, "dbauth", "", "admin", "the database auth db")
	// service coordinates
	rootCmd.PersistentFlags().StringVarP(&a.address, "address", "a", "0.0.0.0", "the http/https listening address")
	rootCmd.PersistentFlags().IntVarP(&a.port, "port", "p", 7443, "the http/https listening port")
	// SSL/TLS
	rootCmd.PersistentFlags().StringVarP(&a.sslCertificate, "certificate", "s", "", "the SSL/TLS certificate PEM file path")
	rootCmd.PersistentFlags().StringVarP(&a.sslPrivateKey, "privatekey", "S", "", "the SSL/TLS private key PEM file path")
	// auth RPC service
	rootCmd.PersistentFlags().StringVarP(&a.authServiceAddress, "authaddr", "A", "", "the authorisation RPC service address")
	rootCmd.PersistentFlags().IntVarP(&a.authServicePort, "authport", "P", 7931, "the authorisation RPC service port")

	// parse flags
	rootCmd.ParseFlags(os.Args)
}

func main() {
	serviceAddress := fmt.Sprintf("%s:%d", a.address, a.port)
	info := make(map[string]string)
	info["bind"] = serviceAddress

	// print logo
	printLogo(info)

	if a.verbose {
		log.MessageLog("Certificate: `%s` | Private key: `%s`\n", a.sslCertificate, a.sslPrivateKey)
	}

	// check for SSL/TLS certificates
	if a.sslCertificate == "" || a.sslPrivateKey == "" {
		log.ErrorLog("Invalid SSL/TLS certificates paths\n")
		os.Exit(1)
	}

	// set up SSL/TLS
	log.MessageLog("Starting SSL/TLS services on address %s.\n", serviceAddress)
	err := http.ListenAndServeTLS(serviceAddress, a.sslCertificate, a.sslPrivateKey, router())
	if err != nil {
		log.ErrorLog("Unable to listen and serve on address %s. cause: [%s].\n", serviceAddress, err.Error())
	}
}

func printLogo(i map[string]string) {
	fmt.Printf("%s", logo.Logo("3n4ChatService", ver.V().VersionString(), i))
}
