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
	"github.com/nexocrew/3nigm4/lib/auth"
)

// Third party libraries
import (
	"github.com/spf13/cobra"
)

// ServeCmd start the http/https server listening
// on exec args, used to register to cobra lib root
// command.
var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve trougth http/https",
	Long:    "Launch http service to expose storage API services.",
	Example: "3n4chatservice serve -d 127.0.0.1:27017 -u dbuser -w dbpwd -a 0.0.0.0 -p 443 -s /tmp/cert.pem -S /tmp/pvkey.pem -v",
}

func init() {
	// service coordinates
	serveCmd.PersistentFlags().StringVarP(&a.address, "address", "a", "0.0.0.0", "the http/https listening address")
	serveCmd.PersistentFlags().IntVarP(&a.port, "port", "p", 7443, "the http/https listening port")
	// SSL/TLS
	serveCmd.PersistentFlags().StringVarP(&a.sslCertificate, "certificate", "s", "", "the SSL/TLS certificate PEM file path")
	serveCmd.PersistentFlags().StringVarP(&a.sslPrivateKey, "privatekey", "S", "", "the SSL/TLS private key PEM file path")
	// files parameters
	serveCmd.RunE = serve
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func serve(cmd *cobra.Command, args []string) error {
	var err error

	cmd.ParseFlags(os.Args)
	serviceAddress := fmt.Sprintf("%s:%d", a.address, a.port)
	info := make(map[string]string)
	info["bind"] = serviceAddress

	if a.verbose {
		log.MessageLog("Certificate: `%s` | Private key: `%s`\n", a.sslCertificate, a.sslPrivateKey)
	}

	authClient, err = auth.NewAuthRPC(a.authServiceAddress, a.authServicePort)
	if err != nil {
		log.ErrorLog("Error creating RPC client connecting to %s:%d. Cause: %s.\n", a.authServiceAddress, a.authServicePort, err.Error())
		os.Exit(1)
	}

	// check for SSL/TLS certificates
	if a.sslCertificate == "" || a.sslPrivateKey == "" {
		log.ErrorLog("Invalid SSL/TLS certificates paths\n")
		os.Exit(1)
	}

	// set up SSL/TLS
	log.MessageLog("Starting SSL/TLS services on address %s.\n", serviceAddress)
	err = http.ListenAndServeTLS(serviceAddress, a.sslCertificate, a.sslPrivateKey, router())
	if err != nil {
		log.ErrorLog("Unable to listen and serve on address %s. cause: [%s].\n", serviceAddress, err.Error())
	}

	return nil
}
