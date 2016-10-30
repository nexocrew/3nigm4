//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Internal dependencies
import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	"github.com/nexocrew/3nigm4/lib/logger"
)

// Third party libs
import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InfoCmd returns file info contained in the argument
// reference file.
var InfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Returns reference file infos",
	Long:    "Retrieves file infos from a locally stored reference file (no interaction with the backend).",
	Example: "3n4cli store info -r /tmp/resources.3rf",
	RunE:    infoReference,
}

const (
	convByte     = 1.0
	convKiloByte = 1024 * convByte
	convMegaByte = 1024 * convKiloByte
	convGigaByte = 1024 * convMegaByte
	convTeraByte = 1024 * convGigaByte
)

// infoReference retrieve and shows infos decrypted from the argument
// reference file.
func infoReference(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)
	// prepare PGP private key
	privateEntityList, err := checkAndLoadPgpPrivateKey(viper.GetString(viperLabel(StoreCmd, "privatekey")))
	if err != nil {
		return err
	}

	// get reference
	refin := viper.GetString(viperLabel(StoreCmd, "referencein"))
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
	// create output logger
	lg := logger.NewLogger(
		color.New(color.BgBlack, color.FgHiWhite),
		"",
		"",
		false,
		true,
	)
	// print out file infos
	lg.Printf("Reference %s infos:\n", refin)
	lg.Printf("\tFile name: %s\n", reference.FileName)
	lg.Printf("\tSize: %.3f Mb\n", float64(reference.Size)/convMegaByte)
	lg.Printf("\tCompressed: %v\n", reference.Compressed)
	lg.Printf("\tChunk size: %.3f Mb\n", float64(reference.ChunkSize)/convMegaByte)
	lg.Printf("\tCheck sum: %s\n", hex.EncodeToString(reference.CheckSum[:]))
	lg.Printf("\tIs directory: %v\n", reference.IsDir)
	lg.Printf("\tLast modification: %s\n", reference.ModTime.String())

	return nil
}
