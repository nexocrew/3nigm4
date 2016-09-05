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
	"io/ioutil"
	"os"
	"path"
)

// Third party libs
import (
	"gopkg.in/yaml.v2"
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

// createPgpKeyPair manage key creation tooltip.
// Golang openpgp library seems not yet provided with a function
// to create encrypted pgp private key files. To maintain the maximum
// level of security is better to delegate the key creation to gpg tool.
// TODO: eventually a command wrapper can be created using golang exec
// package and "gpg --gen-key --openpgp --batch" command.
func createPgpKeyPair() {
	fmt.Printf("***************************************************\n" +
		"Use gpg command to create a new pgp key pair:\n" +
		"\t1. \"gpg --gen-key --openpgp\" to create a new key (use RSA algorith and 4096 bit lenght);\n" +
		"\t2. \"gpg -K\" to list available private keys;\n" +
		"\t3. \"gpg --armor --output ~/.3nigm4/pgp/pvkey.asc --export-secret-keys <key_id>\" to export private key;\n" +
		"\t4. \"gpg --armor --output ~/.3nigm4/pgp/pbkey.asc --export <key_id>\" to export public key.\n" +
		"After exporting the key files verify that ~/.3nigm4/config.yaml have the right reference to key pair files.\n" +
		"\n" +
		"You can also export existing pgp keys from third party services (for ex. Keybase) to be used by 3n4cli. Copy" +
		"them in the ~/.3nigm4/pgp directory.\n" +
		"***************************************************\n")
}

// createDirectories create 3n4cli required dirs.
func createDirectories(rootDir string) error {
	// create it! permission is drwx------
	err := os.Mkdir(rootDir, 0700)
	if err != nil {
		return fmt.Errorf("unable to create %s dir cause %s", rootDir, err.Error())
	}
	// create it! permission is drw-------
	pgpDir := path.Join(rootDir, "pgp")
	err = os.Mkdir(pgpDir, 0700)
	if err != nil {
		return fmt.Errorf("unable to create %s dir cause %s", pgpDir, err.Error())
	}

	log.MessageLog("3n4 directories have been created.\n")
	return nil
}

// initcmd implements initialisation logic.
func initfs(user, rootDir string) error {
	// check for directory presence
	info, err := os.Stat(rootDir)
	if os.IsExist(err) &&
		info.IsDir() {
		return fmt.Errorf("3n4cli root dir %s is already initialised", rootDir)
	}

	// create final structure
	cf := &configFile{}

	// make user set username
	fmt.Printf("Please insert your ususal username (return empty for %s): ", user)
	var username string
	_, err = fmt.Scanln(&username)
	if err != nil {
		return fmt.Errorf("unable to read username input cause %s", err.Error())
	}
	if username != "" {
		cf.Username = username
	} else {
		cf.Username = user
	}

	// make user choose a pgp key
	fmt.Printf("Do you want to use an existing pgp key pair [y,n]: ")
	reader := bufio.NewReader(os.Stdin)
	selection, _, err := reader.ReadRune()
	if err != nil {
		return fmt.Errorf("unable to read selection input cause %s", err.Error())
	}
	switch selection {
	// use an existing key pair passing reference paths
	case 'y':
		fmt.Printf("Insert private pgp key path: ")
		_, err = fmt.Scanln(&cf.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("unable to read private key path input cause %s", err.Error())
		}
		fmt.Printf("Insert public pgp key path: ")
		_, err = fmt.Scanln(&cf.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("unable to read public key path input cause %s", err.Error())
		}
	// create a new key pair
	case 'n':
		createPgpKeyPair()
	default:
		return fmt.Errorf("unknown selection %s expecting \"y\" or \"n\"", selection)
	}

	err = createDirectories(rootDir)
	if err != nil {
		return nil
	}

	// encode the file
	configBinary, err := yaml.Marshal(cf)
	if err != nil {
		return fmt.Errorf("unable to marshal config struct cause %s", err.Error())
	}
	// save to disk
	err = ioutil.WriteFile(path.Join(rootDir, "config.yaml"), configBinary, 0600)
	if err != nil {
		return fmt.Errorf("unable to save config file: %s", err.Error())
	}
	log.MessageLog("3n4 config.yaml has been saved.\n")

	return nil
}
