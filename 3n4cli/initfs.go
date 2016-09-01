//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Std golang libs
import (
	"bufio"
	"fmt"
	"os"
)

// configFile is usable to create the default config
// file on disk
type configFile struct {
	Verbose        bool   `yaml:"verbose,omitempty"`
	MasterKey      bool   `yaml:"masterkey,omitempty"`
	ChunkSize      uint   `yaml:"chunksize,omitempty"`
	PrivateKeyPath string `yaml:"privatekey,omitempty"`
	PublicKeyPath  string `yaml:"publickey,omitempty"`
	Compressed     bool   `yaml:"compressed,omitempty"`
	// services
	StorageServiceAddress string `yaml:"storageaddress,omitempty"`
	StorageServicePort    int    `yaml:"storageport,omitempty"`
	AuthServiceAddress    string `yaml:"authaddress,omitempty"`
	AuthServicePort       int    `yaml:"authport,omitempty"`
	// workers and queues
	Workers int `yaml:"workerscount,omitempty"`
	Queue   int `yaml:"queuesize,omitempty"`
	// identity
	Username string `yaml:"username,omitempty"`
}

// initcmd implements initialisation logic.
func initfs(user, rootDir string) error {
	// check for directory presence
	info, err := os.Stat(rootDir)
	if os.IsExist(err) &&
		info.IsDir() {
		return fmt.Errorf("3n4cli root dir %s is already initialised", rootDir)
	}

	// create it! permission is drwx------
	err = os.Mkdir(rootDir, 0700)
	if err != nil {
		return fmt.Errorf("unable to create %s dir cause %s", rootDir, err.Error())
	}

	// create final structure
	cf := &configFile{}

	// make user set username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please insert your ususal username (return empty for default): ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read username input cause %s", err.Error())
	}
	if username != "" {
		cf.Username = username
	}

	// make user choose a pgp key

	return nil
}
