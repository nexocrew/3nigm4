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
	ct "github.com/nexocrew/3nigm4/lib/commons"
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	sc "github.com/nexocrew/3nigm4/lib/storageclient"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
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
	Example: "3n4cli store upload -k /tmp/userA.asc,/tmp/userB.asc -M -O /tmp/resources.3rf -i ~/file.ext -p 2 -v",
}

func init() {
	StoreCmd.AddCommand(UploadCmd)

	// encryption
	setArgument(UploadCmd, "destkeys")
	setArgument(UploadCmd, "masterkey")
	// i/o paths
	setArgument(UploadCmd, "input")
	setArgument(UploadCmd, "referenceout")
	setArgument(UploadCmd, "chunksize")
	setArgument(UploadCmd, "compressed")
	// working queue setup
	setArgument(UploadCmd, "workerscount")
	setArgument(UploadCmd, "queuesize")
	// resource properties
	setArgument(UploadCmd, "timetolive")
	setArgument(UploadCmd, "permission")
	setArgument(UploadCmd, "sharingusers")

	viper.BindPFlags(UploadCmd.Flags())

	// files parameters
	UploadCmd.RunE = upload
}

// upload send a local file to remote storage after encrypting,
// dividing in chunks, compress and referenced. All these security
// critical operations are done client side only encrypted chunks
// are sent to the server. PGP is used to secure generated reference
// file.
func upload(cmd *cobra.Command, args []string) error {
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// prepare PGP keys
	var entityList openpgp.EntityList
	usersPublicKeys, err := checkAndLoadPgpPublicKey(viper.GetString(am["publickey"].name))
	if err != nil {
		return err
	}
	entityList = append(entityList, usersPublicKeys...)
	recipientsKeys, err := loadRecipientsPublicKeys(viper.GetStringSlice(am["destkeys"].name))
	if err != nil {
		return err
	}
	entityList = append(entityList, recipientsKeys...)
	// get private key
	signerEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(am["privatekey"].name))
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
	if viper.GetBool(am["masterkey"].name) {
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

	// create new encryption chunks
	ec, err := fm.NewEncryptedChunks(
		masterkey,
		viper.GetString(am["input"].name),
		uint64(viper.GetInt(am["chunksize"].name)),
		viper.GetBool(am["compressed"].name))
	if err != nil {
		return err
	}

	// upload resources and get reference file
	rf, err := ec.SaveChunks(
		ds,
		viper.GetDuration(am["timetolive"].name),
		&fm.Permission{
			Permission:   ct.Permission(viper.GetInt(am["permission"].name)),
			SharingUsers: viper.GetStringSlice(am["sharingusers"].name),
		})
	if err != nil {
		return err
	}

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
	destinationPath := viper.GetString(am["referenceout"].name)
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
