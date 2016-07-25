//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package messages

import (
	"time"
)

// RecipientsKys is replicated for each recipient and
// used in handshake flow to exchange, in encrypted form,
// all required symmetric keys.
type RecipientKeys struct {
	Id                   string `json:"id" xml:"id"`             // recipient id;
	EncryptedSessionKeys []byte `json:"sessionk" xml:"sessionk"` // encrypted SessionKeys json encoded struct.
}

// ServerKey is passed to the server will be used to encrypt a
// second random key to be used in symmetric algorithm
// assigning a time to live.
type ServerKey struct {
	ServerSymmetricKey []byte `json:"serverk" xml:"serverk"` // symmetric key to be used to encrypt server random key;
	TimeToLive         uint64 `json:"ttl" xml:"ttl"`         // time to live in seconds.
}

// Handshake request to require a new session to the server
// All the request will be encoded and encrypted using server
// pgp public key.
type HandshakeReq struct {
	TimeStamp      time.Time       `json:"timestamp" xml:"timestamp"`     // handshake op timestamp;
	RecipientsKeys []RecipientKeys `json:"recipientsk" xml:"recipientsk"` // recipient completed with pgp encrypted messages;
	ServerKey      ServerKey       `json:"serverk" xml:"serverk"`         // key that'll be used by the server to encrypt a random generated key;
}

// Handshake successful server response returns all needed
// informations to start exchanging messages with required
// recipients.
type HandshakeRes struct {
	SessionId []byte `json:"session" xml:"session"` // the id of the session.
}

// EnrollmentRes is passed by the server to clients when they
// require pending messages to the service.
type EnrollmentRes struct {
	SessionId            []byte `json:"session" xml:"session"`   // the id of the session;
	EncryptedSessionKeys []byte `json:"sessionk" xml:"sessionk"` // encrypted SessionKeys json encoded struct.
}

// StandardResponse is a generic response message
// used to pass non specific messages like everything
// is OK or an error occurred.
type StandardResponse struct {
	Status string `json:"status"` // Status string
	Error  string `json:"error"`  // Error description
}
