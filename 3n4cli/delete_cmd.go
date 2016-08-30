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
	"os"
)

// Internal dependencies
import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	sc "github.com/nexocrew/3nigm4/lib/storageclient"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DeleteCmd removes remote resources starting from a reference
// file
var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Removes remote resources",
	Long:    "Removes remote resources starting from a reference, it deletes the reference file itself at the end of the process.",
	Example: "3n4cli store delete -r /tmp/resources.3rf -v",
}

func init() {
	StoreCmd.AddCommand(DeleteCmd)

	// i/o paths
	setArgumentPFlags(DeleteCmd, "referencein", &arguments.referenceInPath)
	// working queue setup
	setArgument(DeleteCmd, "workerscount", &arguments.workers)
	setArgument(DeleteCmd, "queuesize", &arguments.queue)

	viper.BindPFlags(DeleteCmd.Flags())

	// files parameters
	DeleteCmd.RunE = deleteReference
}

// deleteReference uses datastorage struct to remotely delete all chunks
// pointed by a reference file.
func deleteReference(cmd *cobra.Command, args []string) error {
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// prepare PGP private key
	privateEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(am["privatekey"].name))
	if err != nil {
		return err
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
	refin := arguments.referenceInPath
	encBytes, err := ioutil.ReadFile(refin)
	if err != nil {
		return fmt.Errorf("unable to access reference file %s cause %s", refin, err.Error())
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

	// delete resources from reference
	err = fm.DeleteChunks(ds, &reference)
	if err != nil {
		return err
	}

	// remove reference file
	os.Remove(refin)

	log.MessageLog("Successfully deleted file.\n")

	return nil
}
