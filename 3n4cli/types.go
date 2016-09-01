//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

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
