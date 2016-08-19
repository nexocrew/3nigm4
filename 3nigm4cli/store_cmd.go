//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	_ "fmt"
	"os/user"
	"path"
)

// Internal dependencies
import ()

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
	Example: "3nigm4cli store",
}

func init() {
	// API references
	StoreCmd.PersistentFlags().StringVarP(&arguments.storageService.Address, "storageaddrs", "", "https://www.nexo.cloud", "the storage service address")
	StoreCmd.PersistentFlags().IntVarP(&arguments.storageService.Port, "storageport", "", 443, "the storage service port")
	// encryption
	StoreCmd.PersistentFlags().StringVarP(&arguments.privateKeyPath, "privkey", "K", "$HOME/.3nigm4/pgp/pvkey.asc", "path for the PGP private key used to decode reference files")

	viper.BindPFlag("StorageServiceAddress", StoreCmd.PersistentFlags().Lookup("storageaddrs"))
	viper.BindPFlag("StorageServicePort", StoreCmd.PersistentFlags().Lookup("storageport"))
	viper.BindPFlag("PgpPrivateKeyPath", StoreCmd.PersistentFlags().Lookup("privkey"))

	viper.SetDefault("StorageServiceAddress", "https://www.nexo.cloud")
	viper.SetDefault("StorageServicePort", 443)
	usr, _ := user.Current()
	viper.SetDefault("PgpPrivateKeyPath", path.Join(usr.HomeDir, ".3nigm4", "pgp", "pvkey.asc"))

	// files parameters
	StoreCmd.RunE = store
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func store(cmd *cobra.Command, args []string) error {
	// load config file
	err := manageConfigFile()
	if err != nil {
		return err
	}

	return nil
}
