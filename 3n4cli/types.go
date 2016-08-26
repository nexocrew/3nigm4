//
// 3nigm4 3nigm4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang packages
import (
	"time"
)

// Third party packages
import (
	"golang.org/x/crypto/openpgp"
)

type pgpKeys struct {
	PrivateKey *openpgp.Entity `json:"-" xml:"-"` // user's in memory private key;
	PublicKey  *openpgp.Entity `json:"-" xml:"-"` // user's in memory public key.
}

type apiService struct {
	Address string `json:"address" xml:"address"`
	Port    int    `json:"port" xml:"port"`
}

// Arguments management struct.
type args struct {
	// server basic args
	verbose   bool
	configDir string
	// login service
	authService apiService
	username    string
	// storage parameters
	storageService apiService
	// data in/out
	inPath           string
	outPath          string
	referenceInPath  string
	referenceOutPath string
	// reference file
	userPrivateKeyPath string
	userPublicKeyPath  string
	publicKeyPaths     []string
	// chunk related
	masterkeyFlag bool
	chunkSize     uint
	compressed    bool
	// workers and queues
	workers int
	queue   int
	// resource properties
	timeToLive   time.Duration
	permission   int
	sharingUsers []string
}
