//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//
package message

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"golang.org/x/crypto/openpgp"
	"io"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

// Struct that contains all required keys to
// participate to a chat session. This structure will be
// encrypted using pgp before being inserted in a Recipient
// keys struct for being sent to the server.
type SessionKeys struct {
	CreatorId          string `json:"creatorid" xml:"creatorid"`   // id of the session creator;
	MainSymmetricKey   []byte `json:"maink" xml:"maink"`           // main random generated symmetric key;
	ServerSymmetricKey []byte `json:"serverk" xml:"serverk"`       // server symmetric key;
	PreSharedFlag      bool   `json:"presharedf" xml:"presharedf"` // is there also a pre-shared key in use;
	PreSharedKey       []byte `json:"-" xml:"-"`                   // pre shared key (only available in the client);
	SessionId          []byte `json:"-" xml:"-"`                   // session id returned by the server after creating the session;
	ServerTmpKey       []byte `json:"-" xml:"-"`                   // server generated in memory key (shoul never be stored anywhere).
}

const (
	kSharedKeySize = 128
)

// randomBytesForLen creates a random data blob
// of length "size".
func randomBytesForLen(size int) ([]byte, error) {
	randData := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, randData); err != nil {
		return nil, err
	}
	return randData, nil
}

// NewSessionKeys creates a new session struct assigning random
// keys and required configurations.
func NewSessionKeys(creatorId string, preshared []byte) (*SessionKeys, error) {
	sk := SessionKeys{}
	var err error
	sk.MainSymmetricKey, err = randomBytesForLen(kSharedKeySize)
	if err != nil {
		return nil, err
	}
	sk.ServerSymmetricKey, err = randomBytesForLen(kSharedKeySize)
	if err != nil {
		return nil, err
	}
	// static values
	sk.CreatorId = creatorId
	if preshared != nil {
		sk.PreSharedFlag = true
		sk.PreSharedKey = preshared
	}
	return &sk, nil
}

// encryptForRecipient creates an encrypted message to pass
// session keys to one of the recipients.
func (sk *SessionKeys) EncryptForRecipients(recipients openpgp.EntityList, signer *openpgp.Entity) ([]byte, error) {
	// encode sessions keys
	encoded, err := json.Marshal(sk)
	if err != nil {
		return nil, err
	}
	// encrypt encoded
	encryped, err := crypto3n.OpenPgpEncrypt(encoded, recipients, signer)
	if err != nil {
		return nil, err
	}
	return encryped, nil
}

func (sk *SessionKeys) EncryptMessage(message []byte) ([]byte, error) {
	mainAesKey := deriveAesKey(sk.MainSymmetricKey)

}

func (sk *SessionKeys) DecryptAndAbsorbeServerKey(chipered []byte) error {

}

func deriveAesKey(longKey []byte) [32]byte {
	return sha256.Sum256(longKey)
}
