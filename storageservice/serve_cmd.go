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
	"net/rpc"
	"strings"
)

// Internal dependencies
import (
	s3c "github.com/nexocrew/3nigm4/lib/s3"
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
	// s3 references
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Endpoint, "s3endpoint", "", "s3.amazonaws.com", "s3 backend service endpoint")
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Region, "s3region", "", "eu-central-1", "s3 backend service region")
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Id, "s3id", "", "", "s3 backend service id")
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Secret, "s3secret", "", "", "s3 backend service secret")
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Token, "s3token", "", "", "s3 backend service token")
	ServeCmd.PersistentFlags().IntVarP(&arguments.s3WorkingQueueSize, "s3wqsize", "", 24, "s3 working queue size")
	ServeCmd.PersistentFlags().IntVarP(&arguments.s3QueueSize, "s3queuesize", "", 300, "s3 queue size")
	ServeCmd.PersistentFlags().StringVarP(&arguments.s3Bucket, "s3bucket", "", "3nigm4", "s3 backend service target bucket")
	// files parameters
	ServeCmd.RunE = serve
}

// Global database referring variable to be copied and released by
// each goroutine.
var db database

// This var is used to permitt to switch to mock db implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses Mongodb
// as backend database system.
var databaseStartup func(*args) (database, error) = mgoStartup

// mgoStartup implement startup logic for a mongodb based database
// connection.
func mgoStartup(a *args) (database, error) {
	// startup db
	mgodb, err := MgoSession(&dbArgs{
		addresses: strings.Split(a.dbAddresses, ","),
		user:      a.dbUsername,
		password:  a.dbPassword,
		authDb:    a.dbAuth,
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
var rpcClient *rpc.Client

// rpcClientStartup creates an RPC client and returns it if no error
// encountered. It manage a success log if all went right.
func rpcClientStartup(a *args) (*rpc.Client, error) {
	address := fmt.Sprintf("%s:%s", a.authServiceAddress, a.authServicePort)
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}

	log.MessageLog("Initialised connection with auth RPC service: %s.\n", address)

	return client, nil
}

// S3 backend service managed with working
// queue.
var s3backend *s3c.S3BackendSession

// s3backendStartup initialise the global s3 backend session.
func s3backendStartup(a *args) (*s3c.S3BackendSession, error) {
	s3, err := s3c.NewS3BackendSession(
		a.s3Endpoint,
		a.s3Region,
		a.s3Id,
		a.s3Secret,
		a.s3Token,
		a.s3WorkingQueueSize,
		a.s3QueueSize,
		a.verbose)
	if err != nil {
		return nil, err
	}

	log.MessageLog("Initialised S3 backend session to endpoint %s on region %s.\n", a.s3Endpoint, a.s3Region)

	return s3, nil
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
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
	rpcClient, err = rpcClientStartup(&arguments)
	if err != nil {
		return fmt.Errorf("unable to start RPC client connection: %s", err.Error())
	}
	defer rpcClient.Close()

	// startup S3 backend
	s3backend, err = s3backendStartup(&arguments)
	if err != nil {
		return fmt.Errorf("unable to initialise s3 backend session: %s", err.Error())
	}
	defer s3backend.Close()
	// start wq chan for async processing
	go manageS3chans(s3backend)

	// create router
	route := mux.NewRouter()
	// define  primary routes
	route.HandleFunc("/sechunk/{id:[A-Fa-f0-9]+}", getChunk).Methods("GET")
	route.HandleFunc("/sechunk/{id:[A-Fa-f0-9]+}", deleteChunk).Methods("DELETE")
	route.HandleFunc("/sechunk", postChunk).Methods("POST")
	route.HandleFunc("/sechunk/{id:[A-Fa-f0-9]+}/verifytx", getVerifyTx).Methods("GET")
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
