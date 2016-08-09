//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// Types shared between different packages, typically used
// to share REST APIs types to be used by server and clients.
//
package commontypes

// Golang std lib.
import (
	"time"
)

// Shared key values
const (
	SecurityTokenKey = "3x4-Security-Token" // header key used to pass security token coded as hexadecimal.
)

// Permission represent the read permission on files.
type Permission int

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

//
// Storage service API structures
//
type SechunkPostRequest struct {
	Id           string        `json:"id"`                // id assigned to the chunk;
	Data         []byte        `json:"data"`              // the data blob to be uploaded;
	TimeToLive   time.Duration `json:"ttl,omitempty"`     // required time to live for the data chunk;
	Permission   Permission    `json:"permission"`        // the type of enforced permission;
	SharingUsers []string      `json:"sharing,omitempty"` // usernames of users enabled to access the file (only in case of Shared permission type).
}
