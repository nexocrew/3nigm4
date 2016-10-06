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
	Type string `bson:"type" json:"type"` // used algorithm.
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

// JobPostRequest body for the POST job API that creates a new
// async job evaluating the value of the "Command" property.
type JobPostRequest struct {
	Command   string            `json:"command"`   // the command type, available commands are: "UPLOAD", "DOWNLOAD", "DELETE";
	Arguments *CommandArguments `json:"arguments"` // command arguments.
}

// CommandArguments argument for various commands
// the presence of some or all the properties depends
// on the invoked command.
type CommandArguments struct {
	// used by all commands
	ResourceID string `json:"resourceid"` // id assigned to the chunk (required for all commands);
	// upload specifics
	Data         []byte        `json:"data,omitempty"`       // the data blob to be uploaded (required for upload);
	TimeToLive   time.Duration `json:"ttl,omitempty"`        // required time to live for the data chunk;
	Permission   Permission    `json:"permission,omitempty"` // the type of enforced permission (required for upload);
	SharingUsers []string      `json:"sharing,omitempty"`    // usernames of users enabled to access the file (only in case of Shared permission type).
}

// JobPostResponse the returned message from the
// pre-flight call, the returned id should be used for
// the verify call.
type JobPostResponse struct {
	JobID string `json:"jobid"` // id for the in processing upload.
}

// JobGetRequest verify API call struct returned to check
// is a required transaction (upload, download, delete) has
// been correctly performed.
type JobGetRequest struct {
	Complete bool     `json:"complete"`           // the upload/download/delete tx has been completed;
	Error    string   `json:"error,omitempty"`    // error description, if any;
	Data     []byte   `json:"data,omitempty"`     // returned requested bytes;
	CheckSum CheckSum `json:"checksum,omitempty"` // data related checksum if any.
}

// OpResult this struct represent the status of an async
// operation, of any type (upload, download, delete, ...).
// Not all field will be present: Error and Data properties
// will be only present if an error occurred or a download
// transaction has been required. Notice that two id are
// managed: a file id (used to identify the target file on S3)
// and RequestID used to associate a request with the async
// result produced. This structure is intended for internal use
// and not to be exposed via APIs.
type OpResult struct {
	ID        string // file id string;
	RequestID string // request (tx) id string (not file id);
	Data      []byte // downloaded data, if any;
	Error     error  // setted if an error was produced fro the upload instruction.
}

// Recipient defines the recipients for a will, this structure is used
// both in REST API handlers and in database representation.
type Recipient struct {
	Name  string `bson:"name" json:"name"`    // recipient's name;
	Email string `bson:"email" json: "email"` // recipient's email address;
	// pgp key identity
	KeyID       uint64 `bson:"keyid" json:"keyid"`             // recipient encryption public key id;
	Fingerprint []byte `bson:"fingerprint" json:"fingerprint"` // recipient encryption public key fingerprint.
}

// WillPostRequest contains all required informations to create
// a new will task from REST APIs.
type WillPostRequest struct {
	// data from reference file
	Reference []byte `json:"reference"` // reference data from storage functionality;
	// will settings
	ExtensionUnit  time.Duration `json:"extensionunit"`  // extension unit to deliver the reference if not delayed by the uploading user;
	NotifyDeadline bool          `json:"notifydeadline"` // notify, via email, to the uploader the deadline approach;
	// sharing settings
	Recipients []Recipient `json:"recipients"` // slice of recipients that shour recieve the reference file.
}

// WillCredentials are used to authenticate the user on will behaviour
// that is enforced returning two quantities (that will be added to typical
// login token): a qrcode and a secondary security code.
type WillCredentials struct {
	QRCode        []byte   `json:"qrcode"`
	SecondaryKeys []string `json:"secondarykeys"`
}

// WillPostResponse returns produced will data from the
// POST handler.
type WillPostResponse struct {
	ID          string           `json:"id"`
	Credentials *WillCredentials `json:"credentials"`
}

// WillGetResponse is the GET response retuned by the handler.
type WillGetResponse struct {
	ID             string        `json:"id"`
	Creation       time.Time     `json:"creation"`
	LastModified   time.Time     `json:"lastmodified"`
	ReferenceFile  []byte        `json:"referencefile"`
	Recipients     []Recipient   `json:"recipients,omitempty"`
	LastPing       time.Time     `json:"lastping"`      // UTC located;
	TimeToDelivery time.Time     `json:"ttd,omitempty"` // UTC located;
	ExtensionUnit  time.Duration `json:"extensionunit,omitempty"`
	NotifyDeadline bool          `json:"notifydeadline,omitempty"`
	DeliveryOffset time.Duration `json:"deliveryoffset,omitempty"`
	Disabled       bool          `json:"disabled,omitempty"`
}

// WillPatchRequest is used to update expiration date for
// a will document.
type WillPatchRequest struct {
	Index        int    `json:"index,omitempty"`
	Otp          string `json:"otp"`
	SecondaryKey string `json:"secondarykey,omitempty"`
}
