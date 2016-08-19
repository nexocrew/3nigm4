//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
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

// UploadCmd can be used to upload a local file to the
// API exposed cloud storage after being divided in chunks and
// encrypted.
var UploadCmd = &cobra.Command{
	Use:     "upload",
	Short:   "Uploads a file to secure storage",
	Long:    "Uploads a local file to the cloud storage returning a resource file usable to retrieve or share data.",
	Example: "3nigm4cli store upload -k /tmp/userA.asc,/tmp/userB.asc -M -o /tmp/resources.3rf --chunksize 3000 --compressed -v",
}

func init() {
	// encryption
	UploadCmd.PersistentFlags().StringSliceVarP(&arguments.publicKeyPaths, "pubkeys", "k", []string{"$HOME/.3nigm4/pgp/pbkey.asc"}, "list of paths of PGP public keys to encode reference files")
	UploadCmd.PersistentFlags().BoolVarP(&arguments.masterkeyFlag, "masterkey", "M", false, "activate the master key insertion")
	// i/o paths
	UploadCmd.PersistentFlags().StringVarP(&arguments.outPath, "output", "o", "$HOME/.3nigm4/references", "directory where output reference files are stored")
	UploadCmd.PersistentFlags().UintVarP(&arguments.chunkSize, "chunksize", "", 1000, "size of encrypted chunks sended to the API frontend")
	UploadCmd.PersistentFlags().BoolVarP(&arguments.compressed, "compressed", "", true, "enable compression of sended data")

	viper.BindPFlag("OutputPath", UploadCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("ChunkSize", UploadCmd.PersistentFlags().Lookup("chunksize"))
	viper.BindPFlag("Compressed", UploadCmd.PersistentFlags().Lookup("compressed"))

	usr, _ := user.Current()
	viper.SetDefault("OutputDir", path.Join(usr.HomeDir, ".3nigm4", "references"))
	viper.SetDefault("ChunkSize", 1000)
	viper.SetDefault("Compressed", true)

	// files parameters
	UploadCmd.RunE = upload
}

func upload(cmd *cobra.Command, args []string) error {
	return nil
}
