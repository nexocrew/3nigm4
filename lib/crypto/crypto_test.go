//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package crypto

import (
	"crypto/sha256"
	"fmt"
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
	key1 = "key12345a293bf93key12345a293bf93"
	key2 = "this.is.a.test.key"
	key3 = "R39eie93oe0903i9e£eoo093"
	salt = "i93ie93e"
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
