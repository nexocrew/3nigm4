//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 04/10/2016
//

package main

// Golang std libs
import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCmd creates a will starting from a resource
// file.
var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Creates and upload a \"will\" activity",
	Long:    "Creates and upload a \"will\" starting from a resource file and a user defined delivery deadline.",
	Example: "3n4cli ishtm create -i ~/reference.3n4",
	PreRun:  verbosePreRunInfos,
}

func init() {
	// i/o paths
	setArgument(CreateCmd, "input")
	setArgument(CreateCmd, "output")
	setArgument(CreateCmd, "extension")
	setArgument(CreateCmd, "notify")
	setArgument(CreateCmd, "recipients")

	viper.BindPFlag(am["input"].name, CreateCmd.PersistentFlags().Lookup(am["input"].name))
	viper.BindPFlag(am["output"].name, CreateCmd.PersistentFlags().Lookup(am["output"].name))
	viper.BindPFlag(am["extension"].name, CreateCmd.PersistentFlags().Lookup(am["extension"].name))
	viper.BindPFlag(am["notify"].name, CreateCmd.PersistentFlags().Lookup(am["notify"].name))
	viper.BindPFlag(am["recipients"].name, CreateCmd.PersistentFlags().Lookup(am["recipients"].name))

	IshtmCmd.AddCommand(CreateCmd)

	// files parameters
	CreateCmd.RunE = create
}

// getRecipients extract components from the recipients
// argument string.
func getRecipients(argument string) []ct.Recipient {
	result := make([]ct.Recipient, 0)
	recipients := strings.Split(argument, ",")
	for _, recipient := range recipients {
		components := strings.Split(recipient, ":")
		// validate components
		if len(components) != 4 {
			continue
		}
		if components[1] == "" {
			continue
		}
		keyid, _ := strconv.Atoi(components[2])
		var signature []byte
		var err error
		signature, err = hex.DecodeString(components[3])
		if err != nil {
			signature = nil
		}
		r := &ct.Recipient{
			Name:        components[0],
			Email:       components[1],
			KeyID:       uint64(keyid),
			Fingerprint: signature,
		}
		result = append(result, *r)
	}
	return result
}

const (
	staticServiceFormatString = "http://%s:%d/v1/ishtm/will"
)

// create create a new "will" resource that will be delivered
// in case something happens to the user.
func create(cmd *cobra.Command, args []string) error {
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// load reference file
	reference, err := ioutil.ReadFile(viper.GetString(am["input"].name))
	if err != nil {
		return fmt.Errorf("unable to open reference file cause %s", err.Error())
	}

	// check for output path
	qrcodePath := viper.GetString(am["output"].name)
	if qrcodePath == "" {
		return fmt.Errorf("a output path is required to save the produced QRCode png image")
	}
	info, err := os.Stat(qrcodePath)
	if os.IsExist(err) {
		return fmt.Errorf("a file named %s already exist, please remove it before proceeding", qrcodePath)
	}
	if info.IsDir() {
		return fmt.Errorf("a file path must be indicated not a directory (%s)", qrcodePath)
	}
	dir := filepath.Dir(qrcodePath)
	info, err = os.Stat(dir)
	if err != nil ||
		info.IsDir() != true {
		return fmt.Errorf("provided path to output file is invalid, %d do not exist", dir)
	}

	// extract recipients
	recipients := getRecipients(viper.GetString(am["recipients"].name))
	if len(recipients) == 0 {
		return fmt.Errorf("unable to create a \"will\" with no recipients")
	}

	// get extension time
	extension := viper.GetInt(am["extension"].name)
	if extension == 0 {
		return fmt.Errorf("unable to create a \"will\" with no extension time")
	}

	willPost := &ct.WillPostRequest{
		Reference:      reference,
		Recipients:     recipients,
		ExtensionUnit:  time.Duration(extension) * time.Minute,
		NotifyDeadline: viper.GetBool("notify"),
	}

	body, err := json.Marshal(willPost)
	if err != nil {
		return fmt.Errorf("unable to marshal request body cause %s", err.Error())
	}

	client := &http.Client{}
	// build request URL
	url := fmt.Sprintf(staticServiceFormatString,
		viper.GetString(am["ishtmeaddress"].name),
		viper.GetInt(am["ishtmport"].name),
	)
	// create will
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to prepare POST requst cause %s", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, pss.Token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform request cause %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code having %d but expected %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body, %s", err.Error())
	}
	var willResponse ct.WillPostResponse
	err = json.Unmarshal(respBody, &willResponse)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response, %s", err.Error())
	}
	resp.Body.Close()

	// save qrcode
	err = ioutil.WriteFile(qrcodePath, willResponse.Credentials.QRCode, 0700)
	if err != nil {
		return fmt.Errorf("critical error unable to save QRCode to file, cause %s, manually save this hex encoded png image",
			err.Error(),
			hex.EncodeToString(willResponse.Credentials.QRCode),
		)
	}
	log.MessageLog("Will ID: %s\n", willResponse.ID)
	log.MessageLog("Will secondary key: %s\n", willResponse.Credentials.SecondaryKey)

	return nil
}
