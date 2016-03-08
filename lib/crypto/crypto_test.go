//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package crypto

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"testing"
)

var (
	plaintex = `This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted.`
	key1          = "key12345a293bf93key12345a293bf93"
	key2          = "this.is.a.test.key"
	key3          = "R39eie93oe0903i9e£eoo093"
	salt          = "i93ie93e"
	pgpPrivateKey = "/home/dyst0ni3/.gnupg/secring.gpg"
	pgpPublicKey  = `-----BEGIN PGP PUBLIC KEY BLOCK-----
	Comment: https://keybase.io/download
	Version: Keybase Go 1.0.13 (darwin)

	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
	=2MuG
	-----END PGP PUBLIC KEY BLOCK-----`
)

func TestXorKeys(t *testing.T) {
	shaedK1 := sha256.Sum256([]byte(key1))
	shaedK2 := sha256.Sum256([]byte(key2))
	shaedK3 := sha256.Sum256([]byte(key3))

	keys := [][]byte{shaedK1[:], shaedK2[:], shaedK3[:]}
	_, err := xorKeys(keys, kRequiredMaxKeySize)
	if err != nil {
		t.Fatalf("Unable to xor keys that's strange: %s.\n", err.Error())
	}

	keys = [][]byte{shaedK1[:], shaedK2[:], []byte(key3)}
	_, err = xorKeys(keys, kRequiredMaxKeySize)
	if err == nil {
		t.Fatalf("In this case xor should return an error key3 is too short.\n")
	}
}

func generateAesKey() ([]byte, error) {
	shaedK1 := sha256.Sum256([]byte(key1))
	shaedK2 := sha256.Sum256([]byte(key2))
	shaedK3 := sha256.Sum256([]byte(key3))

	keys := [][]byte{shaedK1[:], shaedK2[:], shaedK3[:]}
	aesk, err := xorKeys(keys, kRequiredMaxKeySize)
	if err != nil {
		return nil, fmt.Errorf("Unable to xor keys that's strange: %s.\n", err.Error())
	}
	return aesk, nil
}

func TestAESCBCEncryption(t *testing.T) {
	plainbytes := []byte(plaintex)

	// generate keys
	aesk, err := generateAesKey()
	if err != nil {
		t.Fatalf("%s", err)
	}
	derived := DeriveKeyWithPbkdf2(aesk, []byte(salt), 10000)

	// encrypt plain text
	cipheredbytes, err := AesEncrypt(derived, []byte(salt), plainbytes, CBC)
	if err != nil {
		t.Error("Unable to encrypt plain text:", err)
		return
	}
	t.Log("Encryption successfull.")

	// decrypt it
	decryptedbytes, err := AesDecrypt(derived, cipheredbytes, CBC)
	if err != nil {
		t.Error("Unable to decrypt ciphered bytes:", err)
		return
	}
	t.Log("Decryption successfull.")

	if string(decryptedbytes) != plaintex {
		t.Error("Encryption and decryption operations went wrong.")
		return
	}
	t.Log("Correctly AES CBC operations.")
}

func TestOpenPgpEncryption(t *testing.T) {
	pvkr, err := ioutil.ReadFile(pgpPrivateKey)
	if err != nil {
		t.Fatalf("Unable to process %s, it's required to proceed with tests.\n", err.Error())
	}

	plainbytes := []byte(plaintex)
	_, err = OpenPgpEncrypt(plainbytes, pvkr, []byte(pgpPublicKey), "illordlo")
	if err != nil {
		t.Fatalf("Unexpected error: %s.\n", err.Error())
	}
}
