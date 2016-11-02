//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Std golang libs
import (
	"fmt"
	"net/http"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	pingPath = "/v1/ping"
)

// PingCmd let's the check the service status.
var PingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Ping 3n4 services",
	Long:    "Verify that 3n4 services are up, running and available.",
	Example: "3n4cli ping",
	RunE:    ping,
}

// verifyService generically verify if a service ping API respond correctly.
func verifyService(client *http.Client, url, servicename string) error {
	req, err := http.NewRequest(
		"GET",
		url,
		nil)
	if err != nil {
		return err
	}
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.ErrorLog("%s not reachable (status code %d: %s.\n)", servicename, resp.StatusCode, resp.Status)
	}
	log.MessageLog("%s OK.\n", servicename)

	return nil
}

// ping verify the availability of 3n4 services using a specific
// http API.
func ping(cmd *cobra.Command, args []string) error {
	client := &http.Client{}

	// verify Authentication service
	err := verifyService(
		client,
		fmt.Sprintf("%s:%d%s", viper.GetString(viperLabel(cmd, "authaddress")), viper.GetInt(viperLabel(cmd, "authport")), pingPath),
		"Authentication service",
	)
	if err != nil {
		return err
	}

	// verify Storage service
	err = verifyService(
		client,
		fmt.Sprintf("%s:%d%s", viper.GetString(viperLabel(cmd, "storageaddress")), viper.GetInt(viperLabel(cmd, "storageport")), pingPath),
		"Storage service",
	)
	if err != nil {
		return err
	}

	// verify Ishtm service
	err = verifyService(
		client,
		fmt.Sprintf("%s:%d%s", viper.GetString(viperLabel(cmd, "ishtmeaddress")), viper.GetInt(viperLabel(cmd, "ishtmport")), pingPath),
		"Ishtm service",
	)
	if err != nil {
		return err
	}
	return nil
}
