//
// 3nigm4 will package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package will

// Golang std packages
import (
	"bytes"
	"encoding/hex"
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
	// to salt PBKDF2 derivation. In this case is never used cause no PBKDF2
	// is done for db credentials.
	GlobalEncryptionSalt []byte = []byte("00000111")
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

const (
	secondaryKeysNumber = 4
	secondaryKeySize    = 16
)

func generateCredential() (*Credential, []byte, []string, error) {
	// first create
	token, err := hotp.GenerateHOTP(6, true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create hotp cause %s", err.Error())
	}

	// QR code
	qr, err := token.QR("3n4")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to generate QR code cause %s", err.Error())
	}

	tokenEnc, err := encryptHotp(token)
	if err != nil {
		return nil, nil, nil, err
	}

	plainKeys := make([]string, 0)
	secondaryKeys := make([][]byte, 0)
	for idx := 0; idx < secondaryKeysNumber; idx++ {
		authKey, err := ct.RandomBytesForLen(secondaryKeySize)
		if err != nil {
			return nil, nil, nil, err
		}
		// encrypt auth key:
		authKeyEnc, err := crypto3n4.AesEncrypt(
			GlobalEncryptionKey,
			GlobalEncryptionSalt,
			authKey,
			crypto3n4.CBC,
		)
		if err != nil {
			return nil, nil, nil, err
		}
		secondaryKeys = append(secondaryKeys, authKeyEnc)
		plainKeys = append(plainKeys, hex.EncodeToString(authKey))
	}

	return &Credential{
		SecondaryKeys: secondaryKeys,
		SoftwareToken: tokenEnc,
	}, qr, plainKeys, nil
}

func removeKey(slice [][]byte, s int) [][]byte {
	return append(slice[:s], slice[s+1:]...)
}

func verifySecondaryKeys(reference []byte, credentials *Credential) (*Credential, error) {
	if credentials == nil {
		return nil, fmt.Errorf("argument credentials is required and should not be nil")
	}

	var verified bool
	for idx, key := range credentials.SecondaryKeys {
		// decrypt token content
		plaintext, err := crypto3n4.AesDecrypt(
			GlobalEncryptionKey,
			key,
			crypto3n4.CBC,
		)
		if err != nil {
			return nil, err
		}
		// compare
		if bytes.Compare(plaintext, reference) == 0 {
			verified = true
			credentials.SecondaryKeys = removeKey(credentials.SecondaryKeys, idx)
			break
		}
	}

	if verified != true {
		return nil, fmt.Errorf("unknown secondary key")
	}
	return credentials, nil
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
