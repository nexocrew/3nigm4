//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package crypto

// golang standard functions
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

// AesMode defines a enum type for available
// aes encryption modes
type AesMode int8

const (
	CBC AesMode = 0 + iota // AES CBC mode
)

var (
	kSaltSize           = 8  // AES salt size
	kRequiredMaxKeySize = 32 // Max key size (AES256)
	kHmacSha256Size     = 32 // Default size for hmac with sha256
)

// PKCS5Padding padding function to pad a certain
// blob of data with necessary data to be used in
// AES block cipher.
func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// PKCS5UnPadding unpad data after AES block
// decrypting.
func PKCS5UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length <= 0 {
		return nil, fmt.Errorf("Invalid byte blob lenght: expecting > 0 having %d.\n", length)
	}
	unpadding := int(src[length-1])
	delta := length - unpadding
	if delta < 0 {
		return nil, fmt.Errorf("Invalid padding delta lenght: expecting >= 0 having %d.\n", delta)
	}
	return src[:delta], nil
}

// GenerateHMAC produce hmac with a message
// and a key.
func GenerateHMAC(message []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// CheckHMAC verify an hmac message with a given key
// and reference message.
func CheckHMAC(message []byte, messageMAC []byte, key []byte) bool {
	expectedMAC := GenerateHMAC(message, key)
	return hmac.Equal(messageMAC, expectedMAC)
}

// This function derive a key from a password using
// Pbkdf2 algorithm. A good number of iterations is
// ~ 10000 cycles. The derivated key has the right
// lenght for being used in AES256.
func DeriveKeyWithPbkdf2(password []byte, salt []byte, iter int) []byte {
	return pbkdf2.Key(password, salt, iter, kRequiredMaxKeySize, sha1.New)
}

// XOR given keys (passed in a slice)
// returning an unique key.
func xorKeys(keys [][]byte, maxlen int) ([]byte, error) {
	// xor passcodesb
	buffeXored := make([]byte, maxlen)
	for counter, key := range keys {
		if len(key) != maxlen {
			return nil, fmt.Errorf("Invalid passcodes: argument passcodes are too short, should be min 32 byte long.")
		}
		// copy or xor
		if counter == 0 {
			copy(buffeXored, key)
		} else {
			for i := 0; i < maxlen; i++ {
				buffeXored[i] ^= key[i]
			}
		}
	}
	return buffeXored, nil
}

// AesEncrypt encrypt data with AES256 using a key.
// Salt and IV will be passed in the encrypted message.
func AesEncrypt(key []byte, salt []byte, plaintext []byte, mode AesMode, iterations int) ([]byte, error) {
	// check input values
	if len(key) < 1 ||
		len(plaintext) < 1 ||
		plaintext == nil ||
		iterations < 1 {
		return nil, fmt.Errorf("Invalid arguments: passcodes, plaintext and iterations should be not null or empty.")
	}

	// pad plain text
	paddedPlaintext := PKCS5Padding(plaintext, aes.BlockSize)
	// create out buffer
	ciphertext := make([]byte, len(paddedPlaintext)+kSaltSize+aes.BlockSize)
	// copy salt
	if len(salt) != kSaltSize {
		return nil, fmt.Errorf("Invalid salt size, expecting %d having %d.", kSaltSize, len(salt))
	}
	jdx := 0
	for idx := aes.BlockSize; idx < aes.BlockSize+kSaltSize; idx++ {
		ciphertext[idx] = salt[jdx]
		jdx++
	}

	// Should be previously padded
	if len(paddedPlaintext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("Invalid plain text size: should be a multiple of block size.")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// allocate cipher text buffer
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Select cipher mode
	switch {
	case mode == CBC:
		cipherMode := cipher.NewCBCEncrypter(block, iv)
		cipherMode.CryptBlocks(ciphertext[aes.BlockSize+kSaltSize:], paddedPlaintext)
		break
	}

	// composed as iv + salt + data
	return ciphertext, nil
}

// AesDecrypt decrypt data with AES256 using a key
// Salt and IV are passed in the encrypted message.
func AesDecrypt(key []byte, ciphertext []byte, mode AesMode, iterations int) ([]byte, error) {
	// check input values
	if len(key) < 1 ||
		len(ciphertext) < 1 ||
		ciphertext == nil ||
		iterations < 1 {
		return nil, fmt.Errorf("Invalid arguments: passcodes, ciphertext and iterations should be not null or empty.")
	}

	// get packed values
	iv := ciphertext[:aes.BlockSize]
	//salt := ciphertext[aes.BlockSize : aes.BlockSize+kSaltSize]
	ciphert := ciphertext[aes.BlockSize+kSaltSize:]

	// check ciphertext lenght
	if len(ciphert) < aes.BlockSize {
		return nil, fmt.Errorf("Cipher text too short, must be at least longer than block size.")
	}
	if len(ciphert)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("Chiper text have wrong size, should be a block size multiple.")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Select cipher mode
	switch {
	case mode == CBC:
		cipherMode := cipher.NewCBCDecrypter(block, iv)
		cipherMode.CryptBlocks(ciphert, ciphert)
		break
	}
	// unpad data
	unpadded, err := PKCS5UnPadding(ciphert)
	if err != nil {
		return nil, err
	}

	return unpadded, nil
}

func getKeyByEmail(keyring openpgp.EntityList, email string) *openpgp.Entity {
	for _, entity := range keyring {
		for _, ident := range entity.Identities {
			if ident.UserId.Email == email {
				return entity
			}
		}
	}
	return nil
}

func OpenPgpEncrypt(data []byte, privateKr []byte, publicKr []byte, signerEmail string) ([]byte, error) {
	// get private key
	rpv := bytes.NewReader(privateKr)
	privring, err := openpgp.ReadKeyRing(rpv)
	if err != nil {
		return nil, err
	}
	privateKey := getKeyByEmail(privring, signerEmail)
	if privateKey == nil {
		return nil, fmt.Errorf("Private key for user %s not found, unable to sign the message and proceeding.", signerEmail)
	}

	// extract recipients keys
	rpb := bytes.NewReader(publicKr)
	pubring, err := openpgp.ReadArmoredKeyRing(rpb)
	if err != nil {
		return nil, err
	}
	// encrypt message
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, pubring, privateKey, nil, nil)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
