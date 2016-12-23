//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Std golang libs
import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"gopkg.in/yaml.v2"
)

// uploadSettings deines storage upload settings.
type uploadSettings struct {
	ChunkSize  uint `yaml:"chunksize,omitempty"`
	Compressed bool `yaml:"compressed,omitempty"`
}

// storageSettings used to define basic storage
// configuration.
type storageSettings struct {
	StorageServiceAddress string         `yaml:"storageaddress,omitempty"`
	StorageServicePort    int            `yaml:"storageport,omitempty"`
	MasterKey             bool           `yaml:"masterkey,omitempty"`
	PrivateKeyPath        string         `yaml:"privatekey,omitempty"`
	PublicKeyPath         string         `yaml:"publickey,omitempty"`
	Upload                uploadSettings `yaml:"upload,omitempty"`
	// workers and queues
	Workers int `yaml:"workerscount,omitempty"`
	Queue   int `yaml:"queuesize,omitempty"`
}

// ishtmSettings basic ishtm settings.
type ishtmSettings struct {
	IshtmServiceAddress string `yaml:"ishtmeaddress,omitempty"`
	IshtmServicePort    int    `yaml:"ishtmport,omitempty"`
}

// loginSettings login function settings.
type loginSettings struct {
	AuthServiceAddress string `yaml:"authaddress,omitempty"`
	AuthServicePort    int    `yaml:"authport,omitempty"`
	// identity
	Username string `yaml:"username,omitempty"`
}

// logoutSettings addresses used for
// logout.
type logoutSettings struct {
	AuthServiceAddress string `yaml:"authaddress,omitempty"`
	AuthServicePort    int    `yaml:"authport,omitempty"`
}

// pingSettings ping used addresses.
type pingSettings struct {
	AuthServiceAddress    string `yaml:"authaddress,omitempty"`
	AuthServicePort       int    `yaml:"authport,omitempty"`
	StorageServiceAddress string `yaml:"storageaddress,omitempty"`
	StorageServicePort    int    `yaml:"storageport,omitempty"`
	IshtmServiceAddress   string `yaml:"ishtmeaddress,omitempty"`
	IshtmServicePort      int    `yaml:"ishtmport,omitempty"`
}

// configFile is usable to create the default config
// file on disk
type configFile struct {
	Verbose bool `yaml:"verbose,omitempty"`
	// services
	Store  storageSettings `yaml:"store,omitempty"`
	Ishtm  ishtmSettings   `yaml:"ishtm,omitempty"`
	Login  loginSettings   `yaml:"login,omitempty"`
	Logout logoutSettings  `yaml:"logout,omitempty"`
	Ping   pingSettings    `yaml:"ping,omitempty"`
}

// createPgpKeyPair manage key creation tooltip.
// Golang openpgp library seems not yet provided with a function
// to create encrypted pgp private key files. To maintain the maximum
// level of security is better to delegate the key creation to gpg tool.
// TODO: eventually a command wrapper can be created using golang exec
// package and "gpg --gen-key --openpgp --batch" command.
/*
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
}*/

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

// TrimLastChar remove the last character for ReadString function
// that ususally habe '\n' as terminator.
func TrimLastChar(s string) string {
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

var pgpCommand = "Key-Type: RSA\n" +
	"Key-Length: 4096\n" +
	"Subkey-Type: RSA\n" +
	"Subkey-Length: 4096\n" +
	"Name-Real: %s\n" +
	"Name-Comment: %s\n" +
	"Name-Email: %s\n" +
	"Expire-Date: 0\n" +
	"Passphrase: %s\n" +
	"%%pubring %s/.3nigm4/pgp/public.pub\n" +
	"%%secring %s/.3nigm4/pgp/key.sec\n" +
	"%%commit\n" +
	"%%echo done\n"

// https://www.gnupg.org/documentation/manuals/gnupg/Unattended-GPG-key-generation.html
func createPgpKeyPair(name, email, comment, passphrase string) (string, string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", "", err
	}

	// create openpgp command file and save to tmp file
	fileContent := fmt.Sprintf(pgpCommand, name, comment, email, passphrase, usr.HomeDir, usr.HomeDir)

	tmpfile, err := ioutil.TempFile("", "gpgtmp")
	if err != nil {
		return "", "", err
	}
	defer ct.SecureFileWipe(tmpfile)
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(fileContent); err != nil {
		return "", "", err
	}

	// call gpg cli to create the key pair
	cmd := exec.Command("gpg", "--batch", "--gen-key", tmpfile.Name())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("unable to run gpg command: %s (%s)", err.Error(), stderr.String())
	}
	fmt.Printf("%s\n", stdout.String())

	return "", "", nil
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
	// init reader
	reader := bufio.NewReader(os.Stdin)

	// make user set username
	fmt.Printf("Please insert your ususal username (return empty for %s): ", user)
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read username input cause %s", err.Error())
	}
	username = TrimLastChar(username)
	if username != "" {
		cf.Login.Username = username
	} else {
		cf.Login.Username = user
	}

	err = createDirectories(rootDir)
	if err != nil {
		return nil
	}

	// make user choose a pgp key
	fmt.Printf("Do you want to use an existing pgp key pair [y,n]: ")
	selection, _, err := reader.ReadRune()
	if err != nil {
		return fmt.Errorf("unable to read selection input cause %s", err.Error())
	}
	switch selection {
	// use an existing key pair passing reference paths
	case 'y':
		fmt.Printf("Insert private pgp key path: ")
		_, err = fmt.Scanln(&cf.Store.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("unable to read private key path input cause %s", err.Error())
		}
		fmt.Printf("Insert public pgp key path: ")
		_, err = fmt.Scanln(&cf.Store.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("unable to read public key path input cause %s", err.Error())
		}
	// create a new key pair
	case 'n':
		// get pgp key password
		fmt.Printf("Insert a new pgp password: ")
		pwd, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		fmt.Printf("Verify pgp password: ")
		cmpPwd, err := gopass.GetPasswdMasked()
		if err != nil {
			return err
		}
		if bytes.Compare(pwd, cmpPwd) != 0 {
			return fmt.Errorf("inserted password do not match with verified one")
		}
		cf.Store.PrivateKeyPath, cf.Store.PublicKeyPath, err = createPgpKeyPair(username, "n.a.", "n.a.", string(pwd))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown selection %s expecting \"y\" or \"n\"", selection)
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
