//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/auth"
)

// Third party libs
import (
	"github.com/spf13/cobra"
)

// ServeCmd serve via RPC cobra command.
var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve trougth RPC",
	Long:    "Launch RPC service to expose authentication services.",
	Example: "authserver serve -d 127.0.0.1:27017 -u dbuser -w dbpwd -a 0.0.0.0 -p 7931 -v",
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbUsername, "dbuser", "u", "", "the database user name")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbPassword, "dbpwd", "w", "", "the database password")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAuth, "dbauth", "", "admin", "the database auth db")
	ServeCmd.PersistentFlags().StringVarP(&arguments.address, "address", "a", "0.0.0.0", "the RPC listening address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.port, "port", "p", 7931, "the RPC listening port")
	// files parameters
	ServeCmd.RunE = serve
}

// This var is used to permitt to switch to mock db implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses Mongodb
// as backend database system.
var databaseStartup func(*args) (auth.Database, error) = mgoStartup

// mgoStartup implement startup logic for a mongodb based database
// connection.
func mgoStartup(arguments *args) (auth.Database, error) {
	// startup db
	mgodb, err := auth.MgoSession(&auth.DbArgs{
		Addresses: strings.Split(arguments.dbAddresses, ","),
		User:      arguments.dbUsername,
		Password:  arguments.dbPassword,
		AuthDb:    arguments.dbAuth,
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

	// set global db
	auth.SetGlobalDbClient(db)
	defer auth.CloseGlobalDbClient()

	// register RPC calls
	login := new(auth.Login)
	rpc.Register(login)
	sessionauth := new(auth.SessionAuth)
	rpc.Register(sessionauth)

	address := fmt.Sprintf("%s:%d", arguments.address, arguments.port)
	log.MessageLog("Ready to serve via tcp on address %s.\n", address)

	// start listening
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("unable to start rpc service %s", err.Error())
	}

	return http.Serve(listener, nil)
}
