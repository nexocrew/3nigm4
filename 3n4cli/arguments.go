//
// 3nigm4 3n4cli package
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
	ver "github.com/nexocrew/3nigm4/lib/version"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
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
	name        string
	shorthand   string
	value       interface{}
	usage       string
	kind        argType
	pathContent bool
	enigmaExt   bool
}

func flagChanged(flags *flag.FlagSet, key string) bool {
	flag := flags.Lookup(key)
	if flag == nil {
		return false
	}
	return flag.Changed
}

// setArgumentFlags set pflag flags with value contained in the am
// variable map.
func setArgumentFlags(command *cobra.Command, arg cliArguments) {
	switch arg.kind {
	case String:
		command.PersistentFlags().StringP(
			arg.name,
			arg.shorthand,
			arg.value.(string),
			arg.usage,
		)
		if arg.pathContent {
			command.PersistentFlags().SetAnnotation(
				arg.name,
				cobra.BashCompSubdirsInDir,
				[]string{},
			)
		}
		if arg.enigmaExt {
			command.PersistentFlags().SetAnnotation(
				arg.name,
				cobra.BashCompFilenameExt,
				[]string{"3n4"},
			)
		}
	case Int:
		command.PersistentFlags().IntP(
			arg.name,
			arg.shorthand,
			arg.value.(int),
			arg.usage,
		)
	case Uint:
		command.PersistentFlags().UintP(
			arg.name,
			arg.shorthand,
			uint(arg.value.(int)),
			arg.usage,
		)
	case StringSlice:
		command.PersistentFlags().StringSliceP(
			arg.name,
			arg.shorthand,
			arg.value.([]string),
			arg.usage,
		)
	case Bool:
		command.PersistentFlags().BoolP(
			arg.name,
			arg.shorthand,
			arg.value.(bool),
			arg.usage,
		)
	case Duration:
		command.PersistentFlags().DurationP(
			arg.name,
			arg.shorthand,
			time.Duration(arg.value.(int)),
			arg.usage,
		)
	}
}

// viperLabel creates the right viper label for a specified
// command.
func viperLabel(cmd *cobra.Command, name string) string {
	parents := make([]string, 0)
	parents = append(parents, cmd.Use)
	var lastCmd *cobra.Command = cmd
	for {
		selCmd := lastCmd.Parent()
		if selCmd != nil {
			parents = append(parents, selCmd.Use)
			lastCmd = selCmd
			continue
		}
		break
	}
	var result string
	for idx := len(parents) - 2; idx >= 0; idx-- {
		result += parents[idx]
		result += "."
	}
	result += am[name].name

	return result
}

// bindPFlag bind a Viper PFlag with the corresponding
// persistent flag.
func bindPFlag(cmd *cobra.Command, name string) {
	label := viperLabel(cmd, name)
	viper.BindPFlag(
		label,
		cmd.PersistentFlags().Lookup(am[name].name),
	)
}

// setArgument invokes setArgumentPFlags before calling Viper config
// manager to integrate values.
func setArgument(command *cobra.Command, key string) {
	arg, ok := am[key]
	if !ok ||
		command == nil {
		panic("invalid argument required")
	}
	setArgumentFlags(command, arg)
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
	"masterkey": cliArguments{
		name:      "masterkey",
		shorthand: "M",
		value:     false,
		usage:     "activate the master key insertion",
		kind:      Bool,
	},
	"input": cliArguments{
		name:        "input",
		shorthand:   "i",
		value:       "",
		usage:       "file or directory to be stored on the secure cloud or \"ishtm\" service",
		kind:        String,
		pathContent: true,
	},
	"output": cliArguments{
		name:        "output",
		shorthand:   "o",
		value:       "",
		usage:       "directory where output files will be stored",
		kind:        String,
		pathContent: true,
	},
	"referencein": cliArguments{
		name:        "referencein",
		shorthand:   "r",
		value:       "",
		usage:       "reference file path",
		kind:        String,
		pathContent: true,
	},
	"referenceout": cliArguments{
		name:        "referenceout",
		shorthand:   "O",
		value:       "",
		usage:       "reference file output path",
		kind:        String,
		pathContent: true,
	},
	"chunksize": cliArguments{
		name:      "chunksize",
		shorthand: "",
		value:     100000,
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
		name:        "privatekey",
		shorthand:   "K",
		value:       "",
		usage:       "path for the user's PGP private key",
		kind:        String,
		pathContent: true,
	},
	"publickey": cliArguments{
		name:        "publickey",
		shorthand:   "k",
		value:       "",
		usage:       "path for the user's PGP public key",
		kind:        String,
		pathContent: true,
	},
	"destkeys": cliArguments{
		name:      "destkeys",
		shorthand: "",
		value:     "",
		usage:     "path for the PGP public keys of message or resource recipients",
		kind:      String,
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
		value:     "",
		usage:     "if permission is setted to shared (1) the user names passed in this list can have access to the uploaded resource",
		kind:      String,
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
	"ishtmeaddress": cliArguments{
		name:      "ishtmeaddress",
		shorthand: "",
		value:     "https://ishtm.3n4.io",
		usage:     "the \"ishtm\" service APIs address",
		kind:      String,
	},
	"ishtmport": cliArguments{
		name:      "ishtmport",
		shorthand: "",
		value:     443,
		usage:     "the \"ishtm\" service port",
		kind:      Int,
	},
	"extension": cliArguments{
		name:      "extension",
		shorthand: "",
		value:     2880,
		usage:     "the \"ishtm\" extension deadline in minutes",
		kind:      Int,
	},
	"notify": cliArguments{
		name:      "notify",
		shorthand: "",
		value:     true,
		usage:     "enable notification of \"ishtm\" deadline approach",
		kind:      Bool,
	},
	"recipients": cliArguments{
		name:      "recipients",
		shorthand: "",
		value:     "",
		usage:     "list of recipients for the \"ishtm\" service, they should listed as <mail>:<name>:<keyid>:<hex_signature>,... mail address is required",
		kind:      String,
	},
	"id": cliArguments{
		name:      "id",
		shorthand: "D",
		value:     "",
		usage:     "the ID for a \"ishtm will\" record",
		kind:      String,
	},
	"secondary": cliArguments{
		name:      "secondary",
		shorthand: "",
		value:     false,
		usage:     "use secondary key, instead of OTP, to authenticate on a \"ishtm will\" manipulating function",
		kind:      Bool,
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
		pwd, err := gopass.GetPasswdMasked()
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

// storageSettingsDescription helper function to print storage
// settings in verbose mode.
func storageSettingsDescription() string {
	return fmt.Sprintf(
		"\tStorage:\n"+
			"\t\tAddress:%s:%d\n"+
			"\t\tInternal parameters: working queue size %d, queue %d\n"+
			"\t\tChunk parameters: size %d compressed %v\n",
		viper.GetString(viperLabel(StoreCmd, "storageaddress")),
		viper.GetInt(viperLabel(StoreCmd, "storageport")),
		viper.GetInt(viperLabel(StoreCmd, "workerscount")),
		viper.GetInt(viperLabel(StoreCmd, "queuesize")),
		viper.GetInt(viperLabel(UploadCmd, "chunksize")),
		viper.GetBool(viperLabel(UploadCmd, "compressed")),
	)
}

// ishtmSettingsDescription helper function to print ishtm
// settings in verbose mode.
func ishtmSettingsDescription() string {
	return fmt.Sprintf(
		"\tIshtm:\n"+
			"\t\tAddress:%s:%d\n",
		viper.GetString(viperLabel(IshtmCmd, "ishtmeaddress")),
		viper.GetInt(viperLabel(IshtmCmd, "ishtmport")),
	)
}

// authSettingsDescription helper function to print auth
// settings in verbose mode.
func authSettingsDescription() string {
	return fmt.Sprintf(
		"\tAuthentication service:\n"+
			"\t\tLogin:%s:%d username %s\n"+
			"\t\tLogout:%s:%d\n",
		viper.GetString(viperLabel(LoginCmd, "authaddress")),
		viper.GetInt(viperLabel(LoginCmd, "authport")),
		viper.GetString(viperLabel(LoginCmd, "username")),
		viper.GetString(viperLabel(LogoutCmd, "authaddress")),
		viper.GetInt(viperLabel(LogoutCmd, "authport")),
	)
}

// verbosePreRunInfos prints out verbose infos before executing
// commands.
func verbosePreRunInfos(cmd *cobra.Command, args []string) {
	if viper.GetBool(viperLabel(RootCmd, "verbose")) == true {
		log.VerboseLog("Using config file: %s.\n", viper.ConfigFileUsed())
		log.VerboseLog(
			"Context:\n"+
				"\tVersion: %s\n"+
				"%s%s%s\n",
			ver.V().VersionString(),
			authSettingsDescription(),
			storageSettingsDescription(),
			ishtmSettingsDescription(),
		)
	}
}
