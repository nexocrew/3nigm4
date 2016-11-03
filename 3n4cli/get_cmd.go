//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 04/10/2016
//

package main

// Golang std libs
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// GetCmd get a will record (totally or partially)
// from a will ID.
var GetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get and download a \"will\" activity",
	Long:    "Get infos and download a \"will\" activity record.",
	Example: "3n4cli ishtm get --id E44AC9C25D690AF5E44AC9 --output ~/reference.3n4",
	RunE:    get,
}

// formatWillReference creates a printable string from the
// will structure.
func formatWillReference(w *ct.WillGetResponse) string {
	var result string
	result += fmt.Sprintf("Will ID: %s\n", w.ID)
	result += fmt.Sprintf("Creation date: %s\n", w.Creation.String())
	result += fmt.Sprintf("Last ping: %s\n", w.LastPing.String())
	result += fmt.Sprintf("Time to delivery: %s\n", w.TimeToDelivery.String())
	result += fmt.Sprintf("Extension unit: %d min\n", w.ExtensionUnit/time.Minute)
	result += fmt.Sprintf("Notify deadline: %v\n", w.NotifyDeadline)
	result += fmt.Sprintf("Delivery offset: %d min\n", w.DeliveryOffset/time.Minute)
	result += fmt.Sprintf("Disabled: %v\n", w.Disabled)
	for idx, recipient := range w.Recipients {
		result += fmt.Sprintf("Recipien %02d: %s %s %d %s\n",
			idx,
			recipient.Name,
			recipient.Email,
			recipient.KeyID,
			hex.EncodeToString(recipient.Fingerprint),
		)
	}
	return result
}

// get perform a GET request directed to ishtm APIs.
func get(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)
	// check for token presence
	if pss.Token == "" {
		return fmt.Errorf("you are not logged in, please call \"login\" command before invoking any other functionality")
	}

	// validate arguments
	id := viper.GetString(viperLabel(cmd, "id"))
	if id == "" {
		return fmt.Errorf("unable to perform request with empty ID")
	}
	output := viper.GetString(viperLabel(cmd, "output"))
	if output == "" {
		log.WarningLog("Output file is nil retrieved reference file will be returned in stdout.\n")
	} else {
		err := ct.VerifyDestinationPath(output)
		if err != nil {
			return err
		}
	}

	client := &http.Client{}
	// base url
	url := fmt.Sprintf(staticServiceFormatString,
		viper.GetString(viperLabel(IshtmCmd, "ishtmeaddress")),
		viper.GetInt(viperLabel(IshtmCmd, "ishtmport")),
	)
	// get will
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/%s", url, id),
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to prepare GET request, cause %s", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, pss.Token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform GET request, cause %s", err.Error())
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
	var willResponse ct.WillGetResponse
	err = json.Unmarshal(respBody, &willResponse)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response, %s", err.Error())
	}
	resp.Body.Close()

	// produce output
	willString := formatWillReference(&willResponse)
	if output != "" {
		err = ioutil.WriteFile(output, willResponse.ReferenceFile, 0600)
		if err != nil {
			return fmt.Errorf("unable to save reference file to %s path, cause %s", output, err.Error())
		}
		willString += fmt.Sprintf("Reference file: %s\n", output)
	} else {
		willString += fmt.Sprintf("Reference data: %s\n", hex.EncodeToString(willResponse.ReferenceFile))
	}
	log.MessageLog("Will data:\n%s\n", willString)

	return nil
}
