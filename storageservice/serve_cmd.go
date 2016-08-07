//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"fmt"
	"net/http"
	"strings"
)

// Third party libs
import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve trougth http/https",
	Long:    "Launch http service to expose storage API services.",
	Example: "storageservice serve -d 127.0.0.1:27017 -u dbuser -w dbpwd -a 0.0.0.0 -p 443 -s /tmp/cert.pem -S /tmp/pvkey.pem -v",
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbUsername, "dbuser", "u", "", "the database user name")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbPassword, "dbpwd", "w", "", "the database password")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAuth, "dbauth", "", "admin", "the database auth db")
	ServeCmd.PersistentFlags().StringVarP(&arguments.address, "address", "a", "0.0.0.0", "the http/https listening address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.port, "port", "p", 7931, "the http/https listening port")
	ServeCmd.PersistentFlags().StringVarP(&arguments.SslCertificate, "certificate", "s", "", "the SSL/TLS certificate PEM file path")
	ServeCmd.PersistentFlags().StringVarP(&arguments.SslPrivateKey, "privatekey", "S", "", "the SSL/TLS private key PEM file path")
	// files parameters
	ServeCmd.RunE = serve
}

// This var is used to permitt to switch to mock db implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses Mongodb
// as backend database system.
var databaseStartup func(*args) (database, error) = mgoStartup

// mgoStartup implement startup logic for a mongodb based database
// connection.
func mgoStartup(arguments *args) (database, error) {
	// startup db
	mgodb, err := MgoSession(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start db connection cause %s", err.Error())
	}

	log.MessageLog("Mongodb %s successfully connected.\n", arguments.dbAddresses)

	// ensure indexes
	err = mgodb.EnsureMongodbIndexes()
	if err != nil {
		log.WarningLog("Ensuring indexes in Mongodb returned error %s.\n", err.Error())
	}
	return mgodb, nil
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func serve(cmd *cobra.Command, args []string) error {
	printLogo()

	db, err := databaseStartup(&arguments)
	if err != nil {
		return err
	}
	defer db.Close()

	// create router
	route := mux.NewRouter()
	// define  primary routes
	route.HandleFunc("/sechunk/{id:[A-Fa-f0-9]+}", getChunk).Methods("GET")
	route.HandleFunc("/sechunk/{id:[A-Fa-f0-9]+}", getChunk).Methods("DELETE")
	route.HandleFunc("/sechunk", postChunk).Methods("POST")
	// utility routes
	route.HandleFunc("/ping", getPing).Methods("GET")
	// root routes
	http.Handle("/", route)

	serviceAddress := fmt.Sprintf("%s:%d", arguments.address, arguments.port)
	// init http or https connection
	if arguments.SslCertificate != "" &&
		arguments.SslPrivateKey != "" {
		log.MessageLog("Starting listening with TLS on address %s.\n", serviceAddress)
		// set up SSL/TLS
		err = http.ListenAndServeTLS(
			serviceAddress,
			arguments.SslCertificate,
			arguments.SslPrivateKey,
			nil)
		if err != nil {
			return fmt.Errorf("https unable to listen and serve on address: %s cause error: %s", serviceAddress, err.Error())
		}
	} else {
		log.WarningLog("Starting listening on address %s (no SSL/TLS this can produce security risks).\n", serviceAddress)
		// plain https
		err = http.ListenAndServe(serviceAddress, nil)
		if err != nil {
			return fmt.Errorf("http unable to listen and serve on address: %s cause error: %s", serviceAddress, err.Error())
		}
	}
	return nil
}
