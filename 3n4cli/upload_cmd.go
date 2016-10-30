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
	"strings"
	"sync"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
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
	"golang.org/x/crypto/openpgp"
)

// UploadCmd can be used to upload a local file to the
// API exposed cloud storage after being divided in chunks and
// encrypted.
var UploadCmd = &cobra.Command{
	Use:     "upload",
	Short:   "Uploads a file to secure storage",
	Long:    "Uploads a local file to the cloud storage returning a resource file usable to retrieve or share data.",
	Example: "3n4cli store upload --destkeys /tmp/userA.asc,/tmp/userB.asc -M -O /tmp/resources.3rf -i ~/file.ext -p 2 -v",
}

// upload send a local file to remote storage after encrypting,
// dividing in chunks, compress and referenced. All these security
// critical operations are done client side only encrypted chunks
// are sent to the server. PGP is used to secure generated reference
// file.
func upload(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// prepare PGP keys
	var entityList openpgp.EntityList
	usersPublicKeys, err := checkAndLoadPgpPublicKey(viper.GetString(viperLabel(StoreCmd, "publickey")))
	if err != nil {
		return err
	}
	entityList = append(entityList, usersPublicKeys...)
	// manually splits string using the strings.Split function
	// as a workaround the bug (issue #112
	// https://github.com/spf13/viper/issues/112) of the Cobra
	// project.
	destkeys := viper.GetString(viperLabel(cmd, "destkeys"))
	if destkeys != "" {
		destinationKeys := strings.Split(destkeys, ",")
		recipientsKeys, err := loadRecipientsPublicKeys(destinationKeys)
		if err != nil {
			return err
		}
		entityList = append(entityList, recipientsKeys...)
	}

	// get private key
	signerEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(viperLabel(StoreCmd, "privatekey")))
	if err != nil {
		return err
	}
	if len(signerEntityList) == 0 {
		return fmt.Errorf("unexpected private key ring size: the ring is empty")
	}
	// force to select the first private key (if more than one are available)
	signer := signerEntityList[0]

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
		viper.GetInt(viperLabel(StoreCmd, "queuesize")),
	)
	if err != nil {
		return err
	}
	defer ds.Close()
	go manageAsyncErrors(errc)

	// create new encryption chunks
	ec, err := fm.NewEncryptedChunks(
		masterkey,
		viper.GetString(viperLabel(cmd, "input")),
		uint64(viper.GetInt(viperLabel(cmd, "chunksize"))),
		viper.GetBool(viperLabel(cmd, "compressed")),
	)
	if err != nil {
		return err
	}

	// create the multibar container
	// this allows our bars to work together without stomping on one another
	progressBars, _ := multibar.New()
	barProgress := progressBars.MakeBar(100, "uploading")
	// listen in for changes on the progress bars
	go progressBars.Listen()

	var context fm.ContextID
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go progressBarUpdate(&context, ds, barProgress, wg)

	// upload resources and get reference file
	// manually splits string using the strings.Split function
	// as a workaround the bug (issue #112
	// https://github.com/spf13/viper/issues/112) of the Cobra
	// project.
	sharingUsers := strings.Split(viper.GetString(viperLabel(cmd, "sharingusers")), ",")
	rf, err := ec.SaveChunks(
		ds,
		viper.GetDuration(viperLabel(cmd, "timetolive")),
		&fm.Permission{
			Permission:   ct.Permission(viper.GetInt(viperLabel(cmd, "permission"))),
			SharingUsers: sharingUsers,
		},
		&context,
	)
	if err != nil {
		return err
	}
	wg.Wait()

	// encode reference file
	refData, err := json.Marshal(rf)
	if err != nil {
		return fmt.Errorf("unable to encode in json format reference file: %s", err.Error())
	}
	// encrypt reference file
	encryptedData, err := crypto3n.OpenPgpEncrypt(refData, entityList, signer)
	if err != nil {
		return fmt.Errorf("unable to encrypt reference file: %s", err.Error())
	}

	// save tp output file
	destinationPath := viper.GetString(viperLabel(cmd, "referenceout"))
	err = ioutil.WriteFile(
		destinationPath,
		encryptedData,
		0644)
	if err != nil {
		return fmt.Errorf("unable to save reference file to output path %s: %s", destinationPath, err.Error())
	}

	log.MessageLog("Successfully uploaded file.\n")

	return nil
}
