//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Std golang libs
import (
	"fmt"
	"os"
	"os/user"
	"path"
)

// Third party libs
import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InitCmd creates 3n4cli directory and default config files.
var InitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialise the 3n4cli root directory",
	Long:    "Create and initialise the 3n4cli root directory under user's home (only if not already initialised).",
	Example: "3n4cli init",
}

func init() {
	// files parameters
	InitCmd.RunE = initcmd
}

type configFile struct {
	Verbose            bool   `yaml:"verbose"`
	ChunkSize          uint   `yaml:"chunksize"`
	UserPrivateKeyPath string `yaml:"privatekey"`
	UserPublicKeyPath  string `yaml:"publickey"`
	Compressed         bool   `yaml:"compressed"`
	// login service
	authService apiService
	// storage parameters
	storageService apiService
	// workers and queues
	workers int
	queue   int
}

var (
	defaultConfig = nil
)

// initcmd implements initialisation logic.
func initcmd(cmd *cobra.Command, args []string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	rootDir := path.Join(usr.HomeDir, rootAppFolder)
	// check for directory presence
	info, err := os.Stat(rootDir)
	if os.IsExist(err) &&
		info.IsDir() {
		log.CriticalLog()
		return fmt.Errorf("3n4cli root dir %s is already initialised", rootDir)
	}

	// create it! permission is drwx------
	err = os.Mkdir(rootDir, 0700)
	if err != nil {
		return fmt.Errorf("unable to create %s dir cause %s", rootDir, err.Error())
	}

}
