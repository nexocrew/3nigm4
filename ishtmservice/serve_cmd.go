//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"net/http"
	"strings"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	ishtmdb "github.com/nexocrew/3nigm4/lib/ishtm/db"
)

// Third party pkgs
import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(ServeCmd)
}

// ServeCmd start the http/https server listening
// on exec args, used to register to cobra lib root
// command.
var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve trougth http/https",
	Long:    "Launch http service to expose ISHTM API services.",
	Example: "ishtmservice serve -d 127.0.0.1:27017 -u dbuser -w dbpwd -a 0.0.0.0 -p 443 -s /tmp/cert.pem -S /tmp/pvkey.pem -v",
}

func init() {
	// database references
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbUsername, "dbuser", "u", "", "the database user name")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbPassword, "dbpwd", "w", "", "the database password")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAuth, "dbauth", "", "admin", "the database auth db")
	// service coordinates
	ServeCmd.PersistentFlags().StringVarP(&arguments.address, "address", "a", "0.0.0.0", "the http/https listening address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.port, "port", "p", 7443, "the http/https listening port")
	// SSL/TLS
	ServeCmd.PersistentFlags().StringVarP(&arguments.SslCertificate, "certificate", "s", "", "the SSL/TLS certificate PEM file path")
	ServeCmd.PersistentFlags().StringVarP(&arguments.SslPrivateKey, "privatekey", "S", "", "the SSL/TLS private key PEM file path")
	// auth RPC service
	ServeCmd.PersistentFlags().StringVarP(&arguments.authServiceAddress, "authaddr", "A", "", "the authorisation RPC service address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.authServicePort, "authport", "P", 7931, "the authorisation RPC service port")
	// files parameters
	ServeCmd.RunE = serve
}

// Global database referring variable to be copied and released by
// each goroutine.
var db ct.Database

// This var is used to permitt to switch to mock db implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses Mongodb
// as backend database system.
var databaseStartup func(*args) (ct.Database, error) = mgoStartup

// mgoStartup implement startup logic for a mongodb based database
// connection.
func mgoStartup(a *args) (ct.Database, error) {
	// startup db
	mgodb, err := ishtmdb.MgoSession(&ct.DbArgs{
		Addresses: strings.Split(a.dbAddresses, ","),
		User:      a.dbUsername,
		Password:  a.dbPassword,
		AuthDb:    a.dbAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start db connection cause %s", err.Error())
	}

	log.MessageLog("Mongodb %s successfully connected.\n", a.dbAddresses)

	// ensure indexes
	err = mgodb.EnsureMongodbIndexes()
	if err != nil {
		log.WarningLog("Ensuring indexes in Mongodb returned error %s.\n", err.Error())
	}
	return mgodb, nil
}

// RPC client instance, usable by different goroutines simultaneously.
var authClient AuthClient

// This var is used to permitt to switch to mock auth implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses RPC Auth
// service to manage authentication.
var authClientStartup func(*args) (AuthClient, error) = rpcClientStartup

// rpcClientStartup creates an RPC client and returns it if no error
// encountered. It manage a success log if all went right.
func rpcClientStartup(a *args) (AuthClient, error) {
	client, err := NewAuthRpc(arguments.authServiceAddress, arguments.authServicePort)
	if err != nil {
		return nil, err
	}

	log.MessageLog("Initialised connection with auth RPC service: %s:%d.\n",
		arguments.authServiceAddress,
		arguments.authServicePort)

	return client, nil
}

// serve command expose a REST API.
func serve(cmd *cobra.Command, args []string) error {
	printLogo()

	// startup db
	var err error
	db, err = databaseStartup(&arguments)
	if err != nil {
		return err
	}
	defer db.Close()

	// startup RPC auth service
	authClient, err = authClientStartup(&arguments)
	if err != nil {
		return fmt.Errorf("unable to start RPC client connection: %s", err.Error())
	}
	defer authClient.Close()

	// create router
	route := mux.NewRouter()
	// define auth routes
	route.HandleFunc("/v1/authsession", login).Methods("POST")
	route.HandleFunc("/v1/authsession", logout).Methods("DELETE")
	// exposed routes to manage ishtm will
	route.HandleFunc("/v1/ishtm/will", postWill).Methods("POST")
	route.HandleFunc("/v1/ishtm/will/{willid:[A-Fa-f0-9]+}", getWill).Methods("GET")
	route.HandleFunc("/v1/ishtm/will/{willid:[A-Fa-f0-9]+}", patchWill).Methods("PATCH")
	route.HandleFunc("/v1/ishtm/will/{willid:[A-Fa-f0-9]+}", deleteWill).Methods("DELETE")
	// utility routes
	route.HandleFunc("/v1/ping", getPing).Methods("GET")
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
