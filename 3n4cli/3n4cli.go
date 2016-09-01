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
	"os"
	"os/user"
	"path"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/logger"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

// Logger global instance
var log *logger.LogFacility

// Global PGP private key: it's loaded the first time a command, that
// uses it, is invoked. After that remains in memory until the program
// is close.
var pgpPrivateKey openpgp.EntityList

// Global PGP public key: it's loaded the first time a command, that
// uses it, is invoked. After that remains in memory until the program
// is close.
var pgpPublicKey openpgp.EntityList

// rootAppFolder is the the name of the root folder used by the 3nigm4
// app to store config files, stored data, etc... This folder will be
// located under the user $HOME dir.
var rootAppFolder = ".3nigm4"

// RootCmd is the base command used by cobra in the storageservice
// exec.
var RootCmd = &cobra.Command{
	Use:   "3n4cli",
	Short: "CLI client for the 3nigm4 services",
	Long:  "Command line client to access 3nigm4 services, it generally requires a network connection to operate.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	setArgument(RootCmd, "verbose")

	viper.BindPFlag(am["verbose"].name, RootCmd.PersistentFlags().Lookup(am["verbose"].name))
}

// checkRequestStatus check request status and if an anomalous
// response status code is present check for the StandardResponse
// error property.
func checkRequestStatus(httpstatus, expected int, body []byte) error {
	if httpstatus != expected {
		var status ct.StandardResponse
		err := json.Unmarshal(body, &status)
		if err != nil {
			return err
		}
		return fmt.Errorf(
			"service returned wrong status code: having %d expecting %d, cause %s",
			httpstatus,
			expected,
			status.Error)
	}
	return nil
}

func initConfig() {
	usr, err := user.Current()
	if err != nil {
		log.CriticalLog("Unable to access user home dir cause %s.\n", err.Error())
		os.Exit(1)
	}
	// set config file references
	viper.SetConfigName("config")
	viper.AddConfigPath(path.Join(usr.HomeDir, rootAppFolder))

	// set env reader
	viper.SetEnvPrefix("3n4env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.CriticalLog("Unable to read config file: %s.\n", err.Error())
		os.Exit(1)
	}
}

// Execute parsing and execute selected
// command.
func Execute() error {
	// execute actual command
	_, err := RootCmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("3n4cli", false, true)

	// start up storage singleton
	pss = newPersistentStorage()
	if pss == nil {
		log.CriticalLog("Unable to start persistant storage, cannot procede.\n")
		os.Exit(1)
	}

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		pss.save()
		os.Exit(1)
	}

	pss.save()
	os.Exit(0)
}
