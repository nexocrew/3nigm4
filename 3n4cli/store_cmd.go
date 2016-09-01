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
	Use:       "store",
	Short:     "Store securely data to the cloud",
	Long:      "Store and manage secured data to the colud. All the encryption routines are executed on the client only encrypted chunks are sended to the server.",
	Example:   "3n4cli store",
	ValidArgs: []string{"upload", "download", "delete"},
}

func init() {
	// API references
	setArgument(StoreCmd, "storageaddress")
	setArgument(StoreCmd, "storageport")
	// encryption
	setArgument(StoreCmd, "privatekey")
	setArgument(StoreCmd, "publickey")
	setArgument(StoreCmd, "masterkey")
	// working queue setup
	setArgument(StoreCmd, "workerscount")
	setArgument(StoreCmd, "queuesize")
	// i/o paths
	setArgument(StoreCmd, "referencein")

	viper.BindPFlag(am["storageaddress"].name, StoreCmd.PersistentFlags().Lookup(am["storageaddress"].name))
	viper.BindPFlag(am["storageport"].name, StoreCmd.PersistentFlags().Lookup(am["storageport"].name))
	viper.BindPFlag(am["privatekey"].name, StoreCmd.PersistentFlags().Lookup(am["privatekey"].name))
	viper.BindPFlag(am["publickey"].name, StoreCmd.PersistentFlags().Lookup(am["publickey"].name))
	viper.BindPFlag(am["masterkey"].name, StoreCmd.PersistentFlags().Lookup(am["masterkey"].name))
	viper.BindPFlag(am["workerscount"].name, StoreCmd.PersistentFlags().Lookup(am["workerscount"].name))
	viper.BindPFlag(am["queuesize"].name, StoreCmd.PersistentFlags().Lookup(am["queuesize"].name))
	viper.BindPFlag(am["referencein"].name, StoreCmd.PersistentFlags().Lookup(am["referencein"].name))

	RootCmd.AddCommand(StoreCmd)

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
