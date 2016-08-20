//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import ()

// Internal dependencies
import ()

// Third party libs
import (
	"github.com/spf13/cobra"
)

// StoreCmd clinet service that connect to the service API
// to manage sensible data, typically exposed operations
// are upload, download and delete.
var StoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "Store securely data to the cloud",
	Long:    "Store and manage secured data to the colud. All the encryption routines are executed on the client only encrypted chunks are sended to the server.",
	Example: "3nigm4cli store",
}

func init() {
	// API references
	setArgument(StoreCmd, "storageaddress", &arguments.storageService.Address)
	setArgument(StoreCmd, "storageport", &arguments.storageService.Port)
	// encryption
	setArgument(StoreCmd, "privatekey", &arguments.userPrivateKeyPath)
	setArgument(StoreCmd, "publickey", &arguments.userPublicKeyPath)

	// files parameters
	StoreCmd.RunE = store
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func store(cmd *cobra.Command, args []string) error {

	return nil
}
