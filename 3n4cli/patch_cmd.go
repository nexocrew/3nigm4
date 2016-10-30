//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std libs
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"net/http"
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

// PatchCmd patch a will record from a will ID.
var PatchCmd = &cobra.Command{
	Use:     "patch",
	Short:   "Patch a \"will\" activity",
	Long:    "Patch a \"will\" activity record extending it's time to delivedy of an extension unit time.",
	Example: "3n4cli ishtm patch --id E44AC9C25D690AF5E44AC9",
	RunE:    patch,
}

// patch perform a PATCH request directed to ishtm APIs.
func patch(cmd *cobra.Command, args []string) error {
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
	secondaryf := viper.GetBool(viperLabel(cmd, "secondary"))

	patchRequest := &ct.WillPatchRequest{}
	// ask for authentication element
	if secondaryf {
		fmt.Printf("Insert secondary key: ")
		secondary, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		patchRequest.SecondaryKey = string(secondary)
	} else {
		fmt.Printf("Insert OTP: ")
		otp, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		patchRequest.Otp = string(otp)
	}

	body, err := json.Marshal(patchRequest)
	if err != nil {
		return fmt.Errorf("unable to marshal request body cause %s", err.Error())
	}

	client := &http.Client{}
	// base url
	url := fmt.Sprintf(staticServiceFormatString,
		viper.GetString(viperLabel(IshtmCmd, "ishtmeaddress")),
		viper.GetInt(viperLabel(IshtmCmd, "ishtmport")),
	)
	// get will
	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf("%s/%s", url, id),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("unable to prepare PATCH request, cause %s", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, pss.Token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform PATCH request, cause %s", err.Error())
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
	var response ct.StandardResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response, %s", err.Error())
	}
	resp.Body.Close()

	if response.Status != ct.AckResponse {
		return fmt.Errorf("response reported an error: %s", response.Error)
	}
	log.MessageLog("Will record update done correctly.\n")

	return nil
}
