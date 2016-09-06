//
// 3nigm4 storageservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"time"
)

// Internal libs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

const (
	defaultDatabaseName           = "storageservice"
	defaultFilesLogCollectionName = "fileslog"
	defaultAsyncTxCollectionName  = "asynctx"
	envDatabaseName               = "NEXO_FILESLOG_DATABASE"
	envFilesLogCollectionName     = "NEXO_FILESLOG_COLLECTION"
	envAsyncTxCollectionName      = "NEXO_ASYNCTX_COLLECTION"
)

// MaxAsyncTxExistance represent the maximum time
// that an async job can remain pending in the database
// before being automatically deleted.
var MaxAsyncTxExistance = 1 * time.Hour

// These are the available permission types applicable to the stored
// files.
const (
	Private ct.Permission = iota // The file will be accessible only by the uploading user;
	Shared  ct.Permission = iota // it'll be available to the list of users specified by the SharingUsers property;
	Public  ct.Permission = iota // It'll be accessible to everyone (even peolple not registered to the service).
)

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
	// s3 backend
	s3Endpoint         string
	s3Region           string
	s3Id               string
	s3Secret           string
	s3Token            string
	s3WorkingQueueSize int
	s3QueueSize        int
	s3Bucket           string
}
