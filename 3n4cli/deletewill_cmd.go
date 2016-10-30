//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std libs
import (
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"net/http"
	"net/url"
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

// DeleteWillCmd remove a will record from a will ID.
var DeleteWillCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a \"will\" activity",
	Long:    "Delete a \"will\" activity record.",
	Example: "3n4cli ishtm delete --id E44AC9C25D690AF5E44AC9",
	RunE:    deleteWill,
}

// deleteWill perform a DELETE request directed to ishtm APIs.
func deleteWill(cmd *cobra.Command, args []string) error {
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

	// base url
	urlString := fmt.Sprintf(staticServiceFormatString,
		viper.GetString(viperLabel(IshtmCmd, "ishtmeaddress")),
		viper.GetInt(viperLabel(IshtmCmd, "ishtmport")),
	)
	urlString = fmt.Sprintf("%s/%s", urlString, id)
	u, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("invalid url address, %s", err.Error())
	}
	q := u.Query()

	// ask for authentication element
	if secondaryf {
		fmt.Printf("Insert secondary key: ")
		secondary, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		q.Set("secondarykey", string(secondary))
	} else {
		fmt.Printf("Insert OTP: ")
		otp, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		q.Set("otp", string(otp))
	}
	u.RawQuery = q.Encode()

	client := &http.Client{}

	// get will
	req, err := http.NewRequest(
		"DELETE",
		u.String(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to prepare DELETE request, cause %s", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, pss.Token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform DELETE request, cause %s", err.Error())
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
	log.MessageLog("Will record deleted correctly.\n")

	return nil
}
