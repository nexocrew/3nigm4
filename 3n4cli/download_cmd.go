//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Internal dependencies
import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	sc "github.com/nexocrew/3nigm4/lib/storageclient"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DownloadCmd can be used to downlaod from a local reference
// file remotely hosted chunks.
var DownloadCmd = &cobra.Command{
	Use:     "download",
	Short:   "Download a resource",
	Long:    "Downlaod starting from a local reference file remote resources.",
	Example: "3n4cli store download -M -o /tmp/file.ext -r /tmp/resources.3rf -v",
}

func init() {
	StoreCmd.AddCommand(DownloadCmd)

	// encryption
	setArgumentPFlags(DownloadCmd, "masterkey", &arguments.masterkeyFlag)
	// i/o paths
	setArgumentPFlags(DownloadCmd, "referencein", &arguments.referenceInPath)
	setArgument(DownloadCmd, "output", &arguments.outPath)
	// working queue setup
	setArgument(DownloadCmd, "workerscount", &arguments.workers)
	setArgument(DownloadCmd, "queuesize", &arguments.queue)

	viper.BindPFlags(DownloadCmd.Flags())

	// files parameters
	DownloadCmd.RunE = download
}

// download retrieve a previously uploaded resource (divided
// in chunks) from the storage server and recompose it starting
// from the saved reference file.
func download(cmd *cobra.Command, args []string) error {
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// prepare PGP private key
	privateEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(am["privatekey"].name))
	if err != nil {
		return err
	}

	// set master key if any passed
	var masterkey []byte
	if arguments.masterkeyFlag {
		fmt.Printf("Insert master key: ")
		masterkey, err = gopass.GetPasswd()
		if err != nil {
			return err
		}
	}

	// create new store manager
	ds, err, errc := sc.NewStorageClient(
		viper.GetString(am["storageaddress"].name),
		viper.GetInt(am["storageport"].name),
		pss.Token,
		viper.GetInt(am["workerscount"].name),
		viper.GetInt(am["queuesize"].name))
	if err != nil {
		return err
	}
	defer ds.Close()
	go manageAsyncErrors(errc)

	// get reference
	encBytes, err := ioutil.ReadFile(arguments.referenceInPath)
	if err != nil {
		return fmt.Errorf("unable to access reference file: %s", err.Error())
	}
	// decrypt it
	refenceBytes, err := crypto3n.OpenPgpDecrypt(encBytes, privateEntityList)
	if err != nil {
		return fmt.Errorf("unable to decrypt reference file: %s", err.Error())
	}
	// unmarshal it
	var reference fm.ReferenceFile
	err = json.Unmarshal(refenceBytes, &reference)
	if err != nil {
		return fmt.Errorf("unable to decode reference file: %s", err.Error())
	}

	// get resources from reference
	ec, err := fm.LoadChunks(ds, &reference, masterkey)
	if err != nil {
		return err
	}

	// save decoded files
	destinationPath := viper.GetString(am["output"].name)
	err = ec.GetFile(destinationPath)
	if err != nil {
		return fmt.Errorf("unable to save file to output path %s: %s", destinationPath, err.Error())
	}

	log.MessageLog("Successfully downloaded %s file as %s.\n", reference.FileName, destinationPath)

	return nil
}
