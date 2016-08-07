//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std lib.
import (
	"time"
)

// Checksum of the uploaded file for later verify.
type CheckSum struct {
	Hash []byte `bson:"hash"` // checksum of data struct;
	Type string `bson:"type"` // used algorithm.
}

// Owner the file owner.
type Owner struct {
	Username string `bson:"username"`         // the user uploading the file;
	OriginIp string `bson:"ipaddr,omitempty"` // origin upload ip address.
}

// Permission represent the read permission on files.
type Permission int

const (
	Private Permission = iota // The file will be accessible only by the uploading user;
	Shared  Permission = iota // it'll be available to the list of users specified by the SharingUsers property;
	Public  Permission = iota // It'll be accessible to everyone (even peolple not registered to the service).
)

// Acl describes access read rules to the file (writing on the file
// is an exclusive privilege of the uploading user).
type Acl struct {
	Permission   Permission `bson:"permission"`        // the type of enforced permission;
	SharingUsers []string   `bson:"sharing,omitempty"` // usernames of users enabled to access the file (only in case of Shared permission type).
}

// FileLog is used to store all informations about an uploaded file,
// this record will be used, later on, to manage access to the file and
// auto-remove policies.
type FileLog struct {
	Id         string        `bson:"id"`            // the file name assiciated with S3 saved data;
	Size       int           `bson:"size"`          // the size og the uploaded data blob;
	Bucket     string        `bson:"bucket"`        // the S3 bucket where data has been saved;
	CheckSum   CheckSum      `bson:"checksum"`      // checksum for the uploaded data;
	Ownership  Owner         `bson:"ownership"`     // info related to the uploading user;
	Acl        Acl           `bson:"acl"`           // access permissions;
	Creation   time.Time     `bson:"creation_time"` // time of the upload;
	TimeToLive time.Duration `bson:"ttl,omitempty"` // time to live for the uploaded file.
}

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
}

// StandardResponse is a generic response message
// used to pass non specific messages like everything
// is OK or an error occurred.
type StandardResponse struct {
	Status string `json:"status"` // Status string
	Error  string `json:"error"`  // Error description
}

// Response status messages
const (
	AckResponse = "ACK" // Acknowledge message;
	NakResponse = "NAK" // Not acknowledged message.
)
