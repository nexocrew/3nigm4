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
	Example: "3n4cli ishtm create --input ~/reference.3n4 --output ~/qrcode.png --extension 9096 --notify true recep@mail.com:Recep:4738293:E44AC9C25D690AF5,john@mail.com:John:8674859:EFAAD0153EAE6EDE",
	RunE:    create,
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
	staticServiceFormatString = "%s:%d/v1/ishtm/will"
)

// create create a new "will" resource that will be delivered
// in case something happens to the user.
func create(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// load reference file
	reference, err := ioutil.ReadFile(viper.GetString(viperLabel(cmd, "input")))
	if err != nil {
		return fmt.Errorf("unable to open reference file cause %s", err.Error())
	}

	// check for output path
	qrcodePath := viper.GetString(viperLabel(cmd, "output"))
	err = ct.VerifyDestinationPath(qrcodePath)
	if err != nil {
		return err
	}

	// extract recipients
	recipients := getRecipients(viper.GetString(viperLabel(cmd, "recipients")))
	if len(recipients) == 0 {
		return fmt.Errorf("unable to create a \"will\" with no recipients")
	}

	// get extension time
	extension := viper.GetInt(viperLabel(cmd, "extension"))
	if extension == 0 {
		return fmt.Errorf("unable to create a \"will\" with no extension time")
	}

	willPost := &ct.WillPostRequest{
		Reference:      reference,
		Recipients:     recipients,
		ExtensionUnit:  time.Duration(extension) * time.Minute,
		NotifyDeadline: viper.GetBool(viperLabel(cmd, "notify")),
	}

	body, err := json.Marshal(willPost)
	if err != nil {
		return fmt.Errorf("unable to marshal request body cause %s", err.Error())
	}

	client := &http.Client{}
	// build request URL
	url := fmt.Sprintf(staticServiceFormatString,
		viper.GetString(viperLabel(IshtmCmd, "ishtmeaddress")),
		viper.GetInt(viperLabel(IshtmCmd, "ishtmport")),
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

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body, %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code having %d but expected %d: %s",
			resp.StatusCode,
			http.StatusOK,
			string(respBody),
		)
	}

	var willResponse ct.WillPostResponse
	err = json.Unmarshal(respBody, &willResponse)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response, %s", err.Error())
	}
	resp.Body.Close()

	// save qrcode
	err = ioutil.WriteFile(qrcodePath, willResponse.Credentials.QRCode, 0600)
	if err != nil {
		return fmt.Errorf("critical error unable to save QRCode to file, cause %s, manually save this hex encoded png image",
			err.Error(),
			hex.EncodeToString(willResponse.Credentials.QRCode),
		)
	}
	log.MessageLog("Will ID: %s\n", willResponse.ID)
	var secondaryString string
	for idx, key := range willResponse.Credentials.SecondaryKeys {
		secondaryString += fmt.Sprintf("\t%d) %s\n", idx, key)
	}
	log.MessageLog("Will secondary keys:\n%s\n", secondaryString)

	return nil
}
