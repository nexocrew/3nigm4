//
// 3nigm4 messages package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//

package messages

// Std golang lib
import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"time"
)

// Internal imports
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

// SessionKeys contains all required keys to
// participate to a chat session. This structure will be
// encrypted using pgp before being inserted in a Recipient
// keys struct for being sent to the server.
type SessionKeys struct {
	CreatorId          string    `json:"creatorid" xml:"creatorid"`        // id of the session creator;
	MainSymmetricKey   []byte    `json:"maink" xml:"maink"`                // main random generated symmetric key;
	ServerSymmetricKey []byte    `json:"serverk" xml:"serverk"`            // server symmetric key;
	PreSharedFlag      bool      `json:"presharedf" xml:"presharedf"`      // is there also a pre-shared key in use;
	RecipientsIds      []string  `json:"recipientsids" xml:"recipientsid"` // slice of id of recipients and senender (all involved entities);
	PreSharedKey       []byte    `json:"-" xml:"-"`                        // pre shared key (only available in the client);
	SessionId          []byte    `json:"-" xml:"-"`                        // session id returned by the server after creating the session;
	IncrementalCounter uint64    `json:"-" xml:"-"`                        // incremental counter of exchanged messages;
	UserId             string    `json:"-" xml:"-"`                        // the user that is interacting with the session;
	ServerTmpKey       []byte    `json:"-" xml:"-"`                        // server generated in memory key (shoul never be stored anywhere);
	Messages           []Message `json:"-" xml:"-"`                        // in memory plain text messages list associated with the session.
}

// ServerMsg contain the exchange structure used
// to pass the server symmetric key to the server.
type ServerMsg struct {
	ServerSymmetricKey []byte   `json:"serverk" xml:"serverk"`             // symmetric key to be used to encrypt server random key;
	RecipientsIds      []string `json:"recipientsids" xml:"recipientsids"` // resipients ids;
	TimeToLive         uint64   `json:"ttl" xml:"ttl"`                     // time to live in seconds.
}

const (
	kSharedKeySize = 128 // Shared key size.
)

// NewSessionKeys creates a new session struct assigning random
// keys and required configurations.
func NewSessionKeys(creatorId string, preshared []byte, recipients []string) (*SessionKeys, error) {
	sk := SessionKeys{}
	var err error
	sk.MainSymmetricKey, err = ct.RandomBytesForLen(kSharedKeySize)
	if err != nil {
		return nil, err
	}
	sk.ServerSymmetricKey, err = ct.RandomBytesForLen(kSharedKeySize)
	if err != nil {
		return nil, err
	}
	// static values
	sk.CreatorId = creatorId
	if preshared != nil {
		sk.PreSharedFlag = true
		sk.PreSharedKey = preshared
	}
	// set recipients
	sk.RecipientsIds = recipients
	// init message list
	sk.Messages = make([]Message, 0)

	return &sk, nil
}

// SessionFromEncryptedMsg create a new session from an
// encrypted message. Pre-shared key have to be inserted manually.
func SessionFromEncryptedMsg(data []byte, recipientk openpgp.EntityList, preshared []byte) (*SessionKeys, error) {
	// decrypt message
	decrypted, err := crypto3n.OpenPgpDecrypt(data, recipientk)
	if err != nil {
		return nil, err
	}
	var session SessionKeys
	err = json.Unmarshal(decrypted, &session)
	if err != nil {
		return nil, err
	}
	// assign preshared
	if session.PreSharedFlag == true {
		session.PreSharedKey = preshared
	}
	return &session, nil
}

// EncryptForRecipientsHandshake creates an encrypted message to pass
// session keys to one or more of recipients.
func (sk *SessionKeys) EncryptForRecipientsHandshake(recipients openpgp.EntityList, signer *openpgp.Entity) ([]byte, error) {
	// encode sessions keys
	encoded, err := json.Marshal(sk)
	if err != nil {
		return nil, err
	}
	// encrypt encoded
	encrypted, err := crypto3n.OpenPgpEncrypt(encoded, recipients, signer)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// EncryptForServerHandshake create an encrypted packet to exchange
// server encryption key with the server. This function will
// use as recipient the public keys exposed by the server
// entity and will sign the message using the sender key. A
// time to live can be specified to define how many time the
// key should be maintained by the server.
func (sk *SessionKeys) EncryptForServerHandshake(recipients openpgp.EntityList, signer *openpgp.Entity, ttl uint64) ([]byte, error) {
	serverk := ServerMsg{
		ServerSymmetricKey: sk.ServerSymmetricKey,
		RecipientsIds:      sk.RecipientsIds,
		TimeToLive:         ttl,
	}
	// encode in json
	encoded, err := json.Marshal(&serverk)
	if err != nil {
		return nil, err
	}
	// encrypt encoded
	encrypted, err := crypto3n.OpenPgpEncrypt(encoded, recipients, signer)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// Message request for an encrypted message using pre-sared keys.
type Message struct {
	SessionId         []byte    `json:"session" xml:"session"`     // the id of the session;
	SenderId          string    `json:"-" xml:"-"`                 // plain text message sender (in memory);
	EncryptedSenderId []byte    `json:"esenderid" xml:"esenderid"` // encrypted sender id;
	Body              []byte    `json:"-" xml:"-"`                 // plaintext body (in memory);
	EncryptedBody     []byte    `json:"body" xml:"body"`           // the actual encrypted message;
	TimeStamp         time.Time `json:"timestamp" xml:"timestamp"` // message op timestamp;
	Counter           uint64    `json:"counter" xml:"counter"`     // message idx.
}

// SignedMessage wrapping message containing message signature.
type SignedMessage struct {
	Message   Message `json:"message" xml:"message"`     // the message;
	Signature []byte  `json:"signature" xml:"signature"` // signature on json coded message.
}

// xoredKey xor all session keys to obtain the final
// AES key.
func (sk *SessionKeys) xoredKey() ([]byte, error) {
	var keys [][]byte
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
	key, err := crypto3n.XorKeys(keys, 32)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// getKey returns a derived key (using
// PBKDF2) using a given salt.
func (sk *SessionKeys) getKey(salt []byte) ([]byte, error) {
	// get the keys
	key, err := sk.xoredKey()
	if err != nil {
		return nil, err
	}
	// pbkdf2 derivation
	derivedKey := crypto3n.DeriveKeyWithPbkdf2(key, salt, 7)
	return derivedKey, nil
}

// getKeyandAndSalt returns a derived key (using
// PBKDF2) and a random generated salt.
func (sk *SessionKeys) getKeyandAndSalt() ([]byte, []byte, error) {
	// get the keys
	key, err := sk.xoredKey()
	if err != nil {
		return nil, nil, err
	}
	// random salt
	salt, err := ct.RandomBytesForLen(8)
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
func (sk *SessionKeys) EncryptMessage(message []byte, signer *openpgp.Entity) ([]byte, error) {
	if sk.UserId == "" {
		return nil, fmt.Errorf("a valid sender user id is required, right now is nil")
	}
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
	// encrypt sender id
	esenderid, err := crypto3n.AesEncrypt(key, salt, []byte(sk.UserId), crypto3n.CBC)
	if err != nil {
		return nil, err
	}
	// generate message
	msg := Message{
		SessionId:         sk.SessionId,
		EncryptedBody:     encrypted,
		TimeStamp:         time.Now(),
		Counter:           sk.IncrementalCounter + 1,
		EncryptedSenderId: esenderid,
	}
	// create wrapper
	wrapper := SignedMessage{
		Message: msg,
	}
	if signer != nil {
		// marshal msg
		jsonmsg, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		// generate signature
		sign, err := crypto3n.OpenPgpSignMessage(jsonmsg, signer)
		if err != nil {
			return nil, err
		}
		wrapper.Signature = sign
	}

	// encode wrapped message
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// checkSignatures verify signature validity using available
// recipients public keys and the sender id to check the right
// user.
func checkSignatures(signature []byte, msg *Message, senderId string, participants openpgp.EntityList) error {
	// check for signature on the wrapper package
	jsonmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	// get key for sender id
	entity := crypto3n.GetKeyByEmail(participants, senderId)
	if entity == nil {
		return fmt.Errorf("unable to find sender id in participant's keys")
	}
	// check it out
	err = crypto3n.OpenPgpVerifySignature(signature, jsonmsg, entity)
	if err != nil {
		return err
	}
	return nil
}

// DecryptMessage decrypt messages using symmetric
// keys and verifying the signature (if enabled), it
// returns the plaintext message, the message time stamp
// and the message count.
func (sk *SessionKeys) DecryptMessage(chipered []byte, participants openpgp.EntityList, signed bool) (*Message, error) {
	// decode message
	var wrapper SignedMessage
	err := json.Unmarshal(chipered, &wrapper)
	if err != nil {
		return nil, err
	}

	// get salt
	salt, err := crypto3n.GetSaltFromCipherText(wrapper.Message.EncryptedBody)
	if err != nil {
		return nil, err
	}

	// derive key
	key, err := sk.getKey(salt)
	if err != nil {
		return nil, err
	}

	// copy sender before decrypting
	// cause padding will change referenced
	// memory area producing inconsistant signature
	var esenderId []byte
	esenderId = append(esenderId, wrapper.Message.EncryptedSenderId...)
	// decript sender
	senderId, err := crypto3n.AesDecrypt(key, esenderId, crypto3n.CBC)
	if err != nil {
		return nil, err
	}

	// if signature is enabled
	if signed {
		err := checkSignatures(wrapper.Signature, &wrapper.Message, string(senderId), participants)
		if err != nil {
			return nil, err
		}
	}

	// decrypt message
	decrypted, err := crypto3n.AesDecrypt(key, wrapper.Message.EncryptedBody, crypto3n.CBC)
	if err != nil {
		return nil, err
	}

	msg := wrapper.Message
	msg.SenderId = string(senderId)
	msg.Body = decrypted

	// compose plaintext message
	return &msg, nil
}

// deriveAesKey returns a SHAed key from the original
// raw random data.
func deriveAesKey(longKey []byte) [32]byte {
	return sha256.Sum256(longKey)
}
