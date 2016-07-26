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
	ServeCmd.PersistentFlags().StringVarP(&arguments.couchbaseCluster, "dbaddr", "d", "", "the database cluster addres")
	ServeCmd.PersistentFlags().StringVarP(&arguments.couchbaseBucket, "dbbucket", "b", "", "the database target bucket")
	ServeCmd.PersistentFlags().StringVarP(&arguments.couchbaseBucketPwd, "dbbucketpwd", "w", "", "the database bucket password")
	ServeCmd.PersistentFlags().StringVarP(&arguments.address, "address", "a", "0.0.0.0", "the RPC listening address")
	ServeCmd.PersistentFlags().IntVarP(&arguments.port, "port", "9", 7931, "the RPC listening port")
	// files parameters
	ServeCmd.RunE = serve
}

func serve(cmd *cobra.Command, args []string) error {
	printLogo()
	// startup db
	var err error
	arguments.bucket, err = startupCouchbaseConnection(arguments.couchbaseCluster,
		arguments.couchbaseBucket,
		arguments.couchbaseBucketPwd)
	if err != nil {
		return fmt.Errorf("failed to start db connection cause %s", err.Error())
	}
	log.MessageLog("Bucket %s successfully opened on cluster %s.\n",
		arguments.couchbaseBucket,
		arguments.couchbaseCluster)

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

type LoginRequestArg struct {
	Username       string
	HashedPassword []byte
}

type LoginResponseArg struct {
	Token []byte
}

type Login int

func (t *Login) Login(args *LoginRequestArg, response *LoginResponseArg) error {
	return nil
}
