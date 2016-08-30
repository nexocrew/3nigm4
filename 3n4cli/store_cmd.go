//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"os"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StoreCmd clinet service that connect to the service API
// to manage sensible data, typically exposed operations
// are upload, download and delete.
var StoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "Store securely data to the cloud",
	Long:    "Store and manage secured data to the colud. All the encryption routines are executed on the client only encrypted chunks are sended to the server.",
	Example: "3n4cli store",
}

func init() {
	RootCmd.AddCommand(StoreCmd)

	// API references
	setArgument(StoreCmd, "storageaddress", &arguments.storageService.Address)
	setArgument(StoreCmd, "storageport", &arguments.storageService.Port)
	// encryption
	setArgument(StoreCmd, "privatekey", &arguments.userPrivateKeyPath)
	setArgument(StoreCmd, "publickey", &arguments.userPublicKeyPath)

	viper.BindPFlags(StoreCmd.Flags())

	// files parameters
	StoreCmd.RunE = store
}

// manageAsyncErrors is a common function used by the various
// store child commands to manage async returned errors. If
// an error is returned exit is invoked.
func manageAsyncErrors(errc <-chan error) {
	for {
		select {
		case err, _ := <-errc:
			log.CriticalLog("Error encountered: %s.\n", err.Error())
			os.Exit(1)
		}
	}
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func store(cmd *cobra.Command, args []string) error {

	return nil
}
