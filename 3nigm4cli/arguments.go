//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang packages
import (
	"fmt"
	"io/ioutil"
	"time"
)

// Internal packages
import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

// argType identify the available flag types, these types are
// described below.
type argType string

const (
	String      argType = "STRING"      // string flag type;
	StringSlice argType = "STRINGSLICE" // []string flag slice;
	Int         argType = "INT"         // int flag;
	Bool        argType = "BOOL"        // bool flag;
	Uint        argType = "UINT"        // uint flag;
	Duration    argType = "DURATION"    // time.Duration flag.
)

// cliArguments is used to define all available flags with name,
// shorthand, value, usage and kind.
type cliArguments struct {
	name      string
	shorthand string
	value     interface{}
	usage     string
	kind      argType
}

// setArgumentPFlags set pflag flags with value contained in the am
// variable map.
func setArgumentPFlags(command *cobra.Command, key string, destination interface{}) {
	arg, ok := am[key]
	if !ok ||
		command == nil ||
		destination == nil {
		panic("invalid argument required")
	}
	switch arg.kind {
	case String:
		command.PersistentFlags().StringVarP(
			destination.(*string),
			arg.name,
			arg.shorthand,
			arg.value.(string),
			arg.usage)
	case Int:
		command.PersistentFlags().IntVarP(
			destination.(*int),
			arg.name,
			arg.shorthand,
			arg.value.(int),
			arg.usage)
	case Uint:
		command.PersistentFlags().UintVarP(
			destination.(*uint),
			arg.name,
			arg.shorthand,
			uint(arg.value.(int)),
			arg.usage)
	case StringSlice:
		command.PersistentFlags().StringSliceVarP(
			destination.(*[]string),
			arg.name,
			arg.shorthand,
			arg.value.([]string),
			arg.usage)
	case Bool:
		command.PersistentFlags().BoolVarP(
			destination.(*bool),
			arg.name,
			arg.shorthand,
			arg.value.(bool),
			arg.usage)
	case Duration:
		command.PersistentFlags().DurationVarP(
			destination.(*time.Duration),
			arg.name,
			arg.shorthand,
			time.Duration(arg.value.(int)),
			arg.usage)
	}
}

// setArgument invokes setArgumentPFlags before calling Viper config
// manager to integrate values.
func setArgument(command *cobra.Command, key string, destination interface{}) {
	setArgumentPFlags(command, key, destination)
	arg, _ := am[key]
	viper.BindPFlag(arg.name, command.PersistentFlags().Lookup(arg.name))
	viper.SetDefault(arg.name, arg.value)
}

// am is the global arguments map.
var am map[string]cliArguments = map[string]cliArguments{
	"verbose": cliArguments{
		name:      "verbose",
		shorthand: "v",
		value:     false,
		usage:     "activate logging verbosity",
		kind:      Bool,
	},
	"config": cliArguments{
		name:      "config",
		shorthand: "c",
		value:     "$HOME/.3nigm4/",
		usage:     "override default config file directory",
		kind:      String,
	},
	"masterkey": cliArguments{
		name:      "masterkey",
		shorthand: "M",
		value:     false,
		usage:     "activate the master key insertion",
		kind:      Bool,
	},
	"input": cliArguments{
		name:      "input",
		shorthand: "i",
		value:     "",
		usage:     "file or directory to be stored on the secure cloud",
		kind:      String,
	},
	"output": cliArguments{
		name:      "output",
		shorthand: "o",
		value:     "",
		usage:     "directory where output files will be stored",
		kind:      String,
	},
	"referencein": cliArguments{
		name:      "referencein",
		shorthand: "r",
		value:     "",
		usage:     "reference file path",
		kind:      String,
	},
	"referenceout": cliArguments{
		name:      "referenceout",
		shorthand: "O",
		value:     "$HOME/.3nigm4/references",
		usage:     "reference file output path",
		kind:      String,
	},
	"chunksize": cliArguments{
		name:      "chunksize",
		shorthand: "",
		value:     1000,
		usage:     "size of encrypted chunks sended to the API frontend",
		kind:      Uint,
	},
	"compressed": cliArguments{
		name:      "compressed",
		shorthand: "",
		value:     true,
		usage:     "enable compression of sended data",
		kind:      Bool,
	},
	"storageaddress": cliArguments{
		name:      "storageaddress",
		shorthand: "",
		value:     "https://store.3n4.io",
		usage:     "the storage service APIs address",
		kind:      String,
	},
	"storageport": cliArguments{
		name:      "storageport",
		shorthand: "",
		value:     443,
		usage:     "the storage service port",
		kind:      Int,
	},
	"privatekey": cliArguments{
		name:      "privatekey",
		shorthand: "K",
		value:     "$HOME/.3nigm4/pgp/pvkey.asc",
		usage:     "path for the user's PGP private key",
		kind:      String,
	},
	"publickey": cliArguments{
		name:      "publickey",
		shorthand: "k",
		value:     "$HOME/.3nigm4/pgp/pbkey.asc",
		usage:     "path for the user's PGP public key",
		kind:      String,
	},
	"destkeys": cliArguments{
		name:      "destkeys",
		shorthand: "",
		value:     []string{},
		usage:     "path for the PGP public keys of message or resource recipients",
		kind:      StringSlice,
	},
	"workerscount": cliArguments{
		name:      "workerscount",
		shorthand: "W",
		value:     12,
		usage:     "number of workers used to concurrently work on user's requests",
		kind:      Int,
	},
	"queuesize": cliArguments{
		name:      "queuesize",
		shorthand: "Q",
		value:     100,
		usage:     "size of the queue used to store incoming request before being processed by workers, this option can affect ram memory usage",
		kind:      Int,
	},
	"timetolive": cliArguments{
		name:      "timetolive",
		shorthand: "",
		value:     0,
		usage:     "time to live of the uploaded resource in nanoseconds, if zero no time to live is defined",
		kind:      Duration,
	},
	"permission": cliArguments{
		name:      "permission",
		shorthand: "p",
		value:     0,
		usage:     "type of access permission profile for the uploaded resource: 0 is for private (only the uploader can access the resource), 1 for shared, 2 for public (anyone can access it)",
		kind:      Int,
	},
	"sharingusers": cliArguments{
		name:      "sharingusers",
		shorthand: "",
		value:     []string{},
		usage:     "if permission is setted to shared (1) the user names passed in this list can have access to the uploaded resource",
		kind:      StringSlice,
	},
	"authaddress": cliArguments{
		name:      "authaddress",
		shorthand: "",
		value:     "https://store.3n4.io",
		usage:     "the authentication service APIs address",
		kind:      String,
	},
	"authport": cliArguments{
		name:      "authport",
		shorthand: "",
		value:     443,
		usage:     "the authentication service port",
		kind:      Int,
	},
	"username": cliArguments{
		name:      "username",
		shorthand: "u",
		value:     "",
		usage:     "the username for authenticate the service user",
		kind:      String,
	},
}

// checkAndLoadPgpPrivateKey verify if a private key ring has been already
// loaded otherwise loads it, requiring the user for the key ring password
// via cli.
func checkAndLoadPgpPrivateKey(keyfile string) (openpgp.EntityList, error) {
	if pgpPrivateKey == nil ||
		len(pgpPrivateKey) == 0 {
		armoredf, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return nil, fmt.Errorf("unable to access user's private key file: %s", err.Error())
		}
		// get user's password
		fmt.Printf("Insert pgp key password: ")
		pwd, err := gopass.GetPasswd()
		if err != nil {
			return nil, err
		}
		// read armored pgp key ring
		entityl, err := crypto3n.ReadArmoredKeyRing(armoredf, pwd)
		if err != nil {
			return nil, fmt.Errorf("unable to read armored pgp key: %s", err.Error())
		}
		pgpPrivateKey = entityl
	}
	return pgpPrivateKey, nil
}

// checkAndLoadPgpPublicKey verify if a public key ring has been already
// loaded otherwise loads it.
func checkAndLoadPgpPublicKey(keyfile string) (openpgp.EntityList, error) {
	if pgpPublicKey == nil ||
		len(pgpPublicKey) == 0 {
		armoredf, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return nil, fmt.Errorf("unable to access user's public key file: %s", err.Error())
		}

		// read armored pgp key ring
		entityl, err := crypto3n.ReadArmoredKeyRing(armoredf, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to read armored pgp key: %s", err.Error())
		}
		pgpPublicKey = entityl
	}
	return pgpPublicKey, nil
}

// loadRecipientsPublicKeys load from armored key files, passed as arguments,
// the contained public keys and returns an entity list complete of
// all openpgp keys.
func loadRecipientsPublicKeys(keys []string) (openpgp.EntityList, error) {
	var entityList openpgp.EntityList
	for _, key := range keys {
		armoredf, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, fmt.Errorf("unable to access public %s key file: %s", key, err.Error())
		}
		// read armored pgp key ring
		entity, err := crypto3n.ReadArmoredKeyRing(armoredf, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to read armored pgp key %s: %s", key, err.Error())
		}
		entityList = append(entityList, entity...)
	}
	return entityList, nil
}
