//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

// Package commons defines types and utility functions shared between
// all 3nigm4 packages, typically used to share REST APIs types to be
// used by server and clients.
//
package commons

// Golang std lib.
import (
	"time"
)

// Shared key values
const (
	SecurityTokenKey = "3x4-Security-Token" // header key used to pass security token coded as hexadecimal.
)

// CheckSum of the uploaded file for later verify.
type CheckSum struct {
	Hash []byte `bson:"hash" json:"hash"` // checksum of data struct;
	Type string `bson:"type" json:"hash"` // used algorithm.
}

// Permission represent the read permission on files.
type Permission int

// StandardResponse is a generic response message
// used to pass non specific messages like everything
// is OK or an error occurred.
type StandardResponse struct {
	Status string `json:"status"` // status string;
	Error  string `json:"error"`  // error description.
}

// Response status messages
const (
	AckResponse = "ACK" // Acknowledge message;
	NakResponse = "NAK" // Not acknowledged message.
)

// LoginRequest is used by the login API to recieve user's
// credentials, should always be used on an SSL/TLS
// protected session.
type LoginRequest struct {
	Username string `json:"username"` // the user name;
	Password string `json:"password"` // the password used by the user.
}

// LoginResponse is used to respond with a session token
// to a login JSON request.
type LoginResponse struct {
	Token string `json:"token"` // the session token.
}

// LogoutResponse returns the invalidated session token
// to REST API calls.
type LogoutResponse struct {
	Invalidated string `json:"invalidated"` // the invalidated session token.
}

// SechunkPostRequest upload API request struct is used
// in pre-flight API call to pass all needed infos.
type SechunkPostRequest struct {
	ID           string        `json:"id"`                // id assigned to the chunk;
	Data         []byte        `json:"data"`              // the data blob to be uploaded;
	TimeToLive   time.Duration `json:"ttl,omitempty"`     // required time to live for the data chunk;
	Permission   Permission    `json:"permission"`        // the type of enforced permission;
	SharingUsers []string      `json:"sharing,omitempty"` // usernames of users enabled to access the file (only in case of Shared permission type).
}

// SechunkAsyncResponse the returned message from the
// pre-flight call, the returned id should be used for
// the verify call.
type SechunkAsyncResponse struct {
	ID string `json:"id"` // id for the in processing upload.
}

// SechunkTxVerify verify API call struct returned to check
// is a required transaction (upload, download, delete) has
// been correctly performed.
type SechunkTxVerify struct {
	Complete bool     `json:"complete"`           // the upload/download/delete tx has been completed;
	Error    string   `json:"error,omitempty"`    // error description, if any;
	Data     []byte   `json:"data,omitempty"`     // returned requested bytes;
	CheckSum CheckSum `json:"checksum,omitempty"` // data related checksum if any.
}
