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
import ()

// Third party libs
import (
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve trougth RPC",
	Long: `Launch RPC service to expose authentication
		services.`,
	Example: "authserver serve -v",
}

func init() {
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbUsername, "dbuser", "u", "", "the database user name")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbPassword, "dbpwd", "w", "", "the database password")
	ServeCmd.PersistentFlags().StringVarP(&arguments.dbAuth, "dbauth", "", "admin", "the database auth db")
	ServeCmd.PersistentFlags().StringVarP(&arguments.address, "address", "a", "0.0.0.0", "the RPC listening address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.port, "port", "9", 7931, "the RPC listening port")
	// files parameters
	ServeCmd.RunE = serve
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func serve(cmd *cobra.Command, args []string) error {
	printLogo()
	// startup db
	var err error
	arguments.dbclient, err = mgoSession(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})
	if err != nil {
		return fmt.Errorf("failed to start db connection cause %s", err.Error())
	}
	defer arguments.dbclient.Close()

	log.MessageLog("Mongodb %s successfully connected.\n", arguments.dbAddresses)

	// register RPC calls
	login := new(Login)
	rpc.Register(login)
	// start listening
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", arguments.address, arguments.port))
	if err != nil {
		return fmt.Errorf("unable to start rpc service %s", err.Error())
	}
	return http.Serve(listener, nil)
}
