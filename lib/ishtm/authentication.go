//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

// Golang std packages
import (
	"bytes"
	"fmt"
)

// Internal packages
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	crypto3n4 "github.com/nexocrew/3nigm4/lib/crypto"
)

// Third party packages
import (
	"github.com/gokyle/hotp"
)

var (
	// GlobalEncryptionKey setted by the lib including service is used
	// to encrypt data in the database.
	GlobalEncryptionKey []byte
	// GlobalEncryptionSalt setted by the lib including service is used
	// to salt PBKDF2 derivation.
	GlobalEncryptionSalt []byte
)

func encryptHotp(token *hotp.HOTP) ([]byte, error) {
	tokenBytes, err := hotp.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal hotp token, cause %s", err.Error())
	}

	// encrypt seed
	tokenEnc, err := crypto3n4.AesEncrypt(
		GlobalEncryptionKey,
		GlobalEncryptionSalt,
		tokenBytes,
		crypto3n4.CBC,
	)
	if err != nil {
		return nil, err
	}

	return tokenEnc, nil
}

func decryptHotp(encryptedToken []byte) (*hotp.HOTP, error) {
	// decrypt token content
	plaintext, err := crypto3n4.AesDecrypt(
		GlobalEncryptionKey,
		encryptedToken,
		crypto3n4.CBC,
	)
	if err != nil {
		return nil, err
	}

	swtoken, err := hotp.Unmarshal(plaintext)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal token data, cause %s", err.Error())
	}
	return swtoken, nil
}

func generateCredential() (*Credential, []byte, error) {
	// first create
	token, err := hotp.GenerateHOTP(8, true)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create hotp cause %s", err.Error())
	}

	// QR code
	qr, err := token.QR("3n4")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate QR code cause %s", err.Error())
	}

	tokenEnc, err := encryptHotp(token)
	if err != nil {
		return nil, nil, err
	}

	authKey, err := ct.RandomBytesForLen(32)
	if err != nil {
		return nil, nil, err
	}
	// encrypt auth key:
	authKeyEnc, err := crypto3n4.AesEncrypt(
		GlobalEncryptionKey,
		GlobalEncryptionSalt,
		authKey,
		crypto3n4.CBC,
	)
	if err != nil {
		return nil, nil, err
	}

	return &Credential{
		SecondaryKey:  authKeyEnc,
		SoftwareToken: tokenEnc,
	}, qr, nil
}

func verifySecondaryKey(key []byte, credentials *Credential) error {
	if credentials == nil {
		return fmt.Errorf("argument credentials is required and should not be nil")
	}

	// decrypt token content
	plaintext, err := crypto3n4.AesDecrypt(
		GlobalEncryptionKey,
		credentials.SecondaryKey,
		crypto3n4.CBC,
	)
	if err != nil {
		return err
	}

	// compare
	if bytes.Compare(plaintext, key) != 0 {
		return fmt.Errorf("unknown secondary key")
	}
	return nil
}

const (
	ckeckIncrementTolerance = 7 // max number of checks before sw token become disaligned.
)

func verifyOTP(code string, credentials *Credential) (*Credential, error) {
	if credentials == nil {
		return nil, fmt.Errorf("argument credentials is required and should not be nil")
	}

	// decrypt token content
	swtoken, err := decryptHotp(credentials.SoftwareToken)
	if err != nil {
		return nil, err
	}

	// verify value
	if swtoken.Scan(code, ckeckIncrementTolerance) != true {
		return nil, fmt.Errorf("provided code is not valid")
	}

	tokenEnc, err := encryptHotp(swtoken)
	if err != nil {
		return nil, err
	}
	credentials.SoftwareToken = tokenEnc

	return credentials, nil
}
