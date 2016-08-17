//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	_ "fmt"
)

// Internal dependencies
import ()

// Third party libs
import (
	"github.com/spf13/cobra"
)

// StoreCmd clinet service that connect to the service API
// to upload or manage sensible data, operations typically
// exposed are upload, download and delete.
var StoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "Store securely data to the cloud",
	Long:    "Store and manage secured data to the colud. All the encryption routines are executed on the client only encrypted chunks are sended to the server.",
	Example: "3nigm4cli store",
}

func init() {
	// database references
	//StoreCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	//StoreCmd.PersistentFlags().IntVarP(&arguments.port, "port", "p", 7443, "the http/https listening port")
	// files parameters
	StoreCmd.RunE = store
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func store(cmd *cobra.Command, args []string) error {
	return nil
}
