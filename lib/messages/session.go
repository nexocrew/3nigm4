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
	"time"
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
	ServerTmpKey       []byte `json:"-" xml:"-"`                   // server generated in memory key (shoul never be stored anywhere);
	IncrementalCounter uint64 `json:"-" xml:"-"`                   // incremental counter of exchanged messages;
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

// Request for an encrypted message using pre-sared keys.
type Message struct {
	SessionId     []byte    `json:"session" xml:"session"`     // the id of the session;
	EncryptedBody []byte    `json:"body" xml:"body"`           // the actual encrypted message;
	TimeStamp     time.Time `json:"timestamp" xml:"timestamp"` // message op timestamp;
	Counter       uint64    `json:"counter" xml:"counter"`     // message idx.
}

type SignedMessage struct {
	Message   Message `json:"message" xml:"message"`     // the message;
	Signature []byte  `json:"signature" xml:"signature"` // signature on json coded message.
}

// getKeyandAndSalt returns a derived key (using
// PBKDF2) and a random generated salt.
func (sk *SessionKeys) getKeyandAndSalt() ([]byte, []byte, error) {
	keys := make([][]byte, 0)
	// main key
	mainAesKey := deriveAesKey(sk.MainSymmetricKey)
	keys = append(keys, mainAesKey[:])
	// server key
	serverAesKey := deriveAesKey(sk.ServerTmpKey)
	keys = append(keys, serverAesKey[:])
	// pre shared
	if sk.PreSharedFlag == true {
		presharedAesKey := deriveAesKey(sk.PreSharedKey)
		keys = append(keys, presharedAesKey[:])
	}
	// get the keys
	key, err := crypto3n.XorKeysKeys(keys, 32)
	if err != nil {
		return nil, nil, err
	}
	// random salt
	salt, err := randomBytesForLen(8)
	if err != nil {
		return nil, nil, err
	}
	// pbkdf2 derivation
	derivedKey := crypto3n.DeriveKeyWithPbkdf2(key, salt, 7)
	return derivedKey, salt, nil
}

// EncryptMessage derive all required keys and encrypt a message using
// pre-shared keys. Notice that server maintained key has been already
// retrieved and is stored in volatile RAM for usage.
func (sk *SessionKeys) EncryptMessage(message []byte) ([]byte, error) {
	// get key and salt
	key, salt, err := sk.getKeyandAndSalt()
	if err != nil {
		return nil, err
	}
	// encrypt it!
	encrypted, err := crypto3n.AesEncrypt(key, salt, message, crypto3n.CBC)
	if err != nil {
		return nil, err
	}
	// generate message
	msg := Message{
		SessionId:     sk.SessionId,
		EncryptedBody: encrypted,
		TimeStamp:     time.Now(),
		Counter:       sk.IncrementalCounter + 1,
	}
	// create wrapper
	wrapper := SignedMessage{
		Message: msg,
	}
	// generate signature

	return key, nil
}

func (sk *SessionKeys) DecryptAndAbsorbeServerKey(chipered []byte) error {
	return nil
}

func deriveAesKey(longKey []byte) [32]byte {
	return sha256.Sum256(longKey)
}
