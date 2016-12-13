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
	"sync"
)

// Internal dependencies
import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	sc "github.com/nexocrew/3nigm4/lib/storageclient"
)

// Third party libs
import (
	"github.com/sethgrid/multibar"
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
	RunE:    deleteReference,
}

// deleteReference uses datastorage struct to remotely delete all chunks
// pointed by a reference file.
func deleteReference(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// prepare PGP private key
	privateEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(viperLabel(StoreCmd, "privatekey")))
	if err != nil {
		return err
	}

	// create new store manager
	ds, err, errc := sc.NewStorageClient(
		viper.GetString(viperLabel(StoreCmd, "storageaddress")),
		viper.GetInt(viperLabel(StoreCmd, "storageport")),
		pss.Token,
		viper.GetInt(viperLabel(StoreCmd, "workerscount")),
		viper.GetInt(viperLabel(StoreCmd, "queuesize")))
	if err != nil {
		return err
	}
	defer ds.Close()
	go manageAsyncErrors(errc)

	// get reference
	refin := viper.GetString(viperLabel(cmd, "referencein"))
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

	// create the multibar container
	// this allows our bars to work together without stomping on one another
	progressBars, _ := multibar.New()
	barProgress := progressBars.MakeBar(100, "deleting")
	// listen in for changes on the progress bars
	go progressBars.Listen()

	var context fm.ContextID
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go progressBarUpdate(&context, ds, barProgress, wg)

	// delete resources from reference
	err = fm.DeleteChunks(ds, &reference, &context)
	if err != nil {
		return err
	}
	wg.Wait()

	// remove reference file
	os.Remove(refin)

	log.MessageLog("Successfully deleted file.\n")

	return nil
}
