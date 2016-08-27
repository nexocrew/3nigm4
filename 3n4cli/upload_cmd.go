//
// 3nigm4 3nigm4cli package
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
	Example: "3nigm4cli store upload -k /tmp/userA.asc,/tmp/userB.asc -M -O /tmp/resources.3rf --chunksize 3000 --compressed -v",
}

func init() {
	// encryption
	setArgument(UploadCmd, "destkeys", &arguments.publicKeyPaths)
	setArgumentPFlags(UploadCmd, "masterkey", &arguments.masterkeyFlag)
	// i/o paths
	setArgumentPFlags(UploadCmd, "input", &arguments.inPath)
	setArgumentPFlags(UploadCmd, "referenceout", &arguments.referenceOutPath)
	setArgument(UploadCmd, "chunksize", &arguments.chunkSize)
	setArgument(UploadCmd, "compressed", &arguments.compressed)
	// working queue setup
	setArgument(UploadCmd, "workerscount", &arguments.workers)
	setArgument(UploadCmd, "queuesize", &arguments.queue)
	// resource properties
	setArgumentPFlags(UploadCmd, "timetolive", &arguments.timeToLive)
	setArgumentPFlags(UploadCmd, "permission", &arguments.permission)
	setArgumentPFlags(UploadCmd, "sharingusers", &arguments.sharingUsers)

	// files parameters
	UploadCmd.RunE = upload
}

// upload send a local file to remote storage after encrypting,
// dividing in chunks, compress and referenced. All these security
// critical operations are done client side only encrypted chunks
// are sent to the server. PGP is used to secure generated reference
// file.
func upload(cmd *cobra.Command, args []string) error {
	// load config file
	err := manageConfigFile()
	if err != nil {
		return err
	}

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

	// create new encryption chunks
	ec, err := fm.NewEncryptedChunks(
		masterkey,
		arguments.inPath,
		uint64(viper.GetInt(am["chunksize"].name)),
		viper.GetBool(am["compressed"].name))
	if err != nil {
		return err
	}

	// upload resources and get reference file
	rf, err := ec.SaveChunks(
		ds,
		arguments.timeToLive,
		&fm.Permission{
			Permission:   ct.Permission(arguments.permission),
			SharingUsers: arguments.sharingUsers,
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
	destinationPath := arguments.referenceOutPath
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
