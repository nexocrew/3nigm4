//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package main

const (
	basePath = "/{version}"
)

// Arguments management struct.
type args struct {
	// server basic args
	verbose bool
	// bind address
	address string
	port    int
	// mongodb address
	dbAddresses string
	dbUsername  string
	dbPassword  string
	dbAuth      string
	// https certificates
	sslCertificate string
	sslPrivateKey  string
	// rpc authservice
	authServiceAddress string
	authServicePort    int
	// workers and queues
	workers int
	queue   int
}
