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
	"github.com/howeyc/gopass"
	"github.com/sethgrid/multibar"
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
	RunE:    download,
}

// download retrieve a previously uploaded resource (divided
// in chunks) from the storage server and recompose it starting
// from the saved reference file.
func download(cmd *cobra.Command, args []string) error {
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

	// set master key if any passed
	var masterkey []byte
	if viper.GetBool(viperLabel(StoreCmd, "masterkey")) {
		fmt.Printf("Insert master key: ")
		masterkey, err = gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
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
	encBytes, err := ioutil.ReadFile(viper.GetString(viperLabel(cmd, "referencein")))
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

	// create the multibar container
	// this allows our bars to work together without stomping on one another
	progressBars, _ := multibar.New()
	barProgress := progressBars.MakeBar(100, "downloading")
	// listen in for changes on the progress bars
	go progressBars.Listen()

	var context fm.ContextID
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go progressBarUpdate(&context, ds, barProgress, wg)

	// get resources from reference
	ec, err := fm.LoadChunks(ds, &reference, masterkey, &context)
	if err != nil {
		return err
	}
	wg.Wait()

	// save decoded files
	destinationPath := viper.GetString(viperLabel(cmd, "output"))
	err = ec.GetFile(destinationPath)
	if err != nil {
		return fmt.Errorf("unable to save file to output path %s: %s", destinationPath, err.Error())
	}

	log.MessageLog("Successfully downloaded %s file as %s.\n", reference.FileName, destinationPath)

	return nil
}
