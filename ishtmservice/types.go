//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

// Arguments management struct.
type args struct {
	// server basic args
	verbose bool
	colored bool
	// mongodb
	dbAddresses string
	dbUsername  string
	dbPassword  string
	dbAuth      string
	// service
	address string
	port    int
	// https
	SslCertificate string
	SslPrivateKey  string
	// auth rpc service
	authServiceAddress string
	authServicePort    int
	// encryption keys
	encryptionKey string
}
