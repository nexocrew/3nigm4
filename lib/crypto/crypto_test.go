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
	// pub   1024R/7F98BBCE 2014-01-04
	// uid                  Golang Test (Private key password is 'golang') <golangtest@test.com>
	// sub   1024R/5F34A320 2014-01-04
	publicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
	Version: GnuPG v1.4.15 (Darwin)
	mI0EUsdthgEEAMKOAoeY+bAHEjdzcM9WhJ27T4QmX8SxLYcRo3rd2cuQawwCz7jf
	bCzLCYyvMqoIvjSxuElVgFx97RyEv5yvLg7ngNfv6ADRlJXMVQ3YQahyeRofPJJ+
	5S0F0JOahZlkAYWIHCUhLtoT/zpI7IeSwWjwtEL1b8YhZBLY9txp29TLABEBAAG0
	REdvbGFuZyBUZXN0IChQcml2YXRlIGtleSBwYXNzd29yZCBpcyAnZ29sYW5nJykg
	PGdvbGFuZ3Rlc3RAdGVzdC5jb20+iLgEEwECACIFAlLHbYYCGwMGCwkIBwMCBhUI
	AgkKCwQWAgMBAh4BAheAAAoJEFVKIId/mLvOookD+wVQzZN8vZVkpYLsTU3XDBly
	0H0F/vtJ4A9JWkYJnRyJRggV3DAajAq2OgOuxtiA+n5QY7JgwPq0bNYpomtBCgPJ
	pCpVVGFs1cHsnPslPZqoocPW3tzHkV9TMMwE2i7dM5YeiYNfJAYMBQsBmeNo6Pz+
	kN7qmjHIGW5KMwlTN8OmuI0EUsdthgEEAK1DA6pBp4PQqaZO91AVgXe44YW7ZNHm
	kUIf4KFB4SiXq2eCzENtSCsiF/hkG7HA6XHKVzCOnk4V8ay/g/BuHDW+HsL09M3N
	tPk/dc7YE/QP+FYn3BD0AhK06mP6GaYQM2TNaerEXp3NtnuNok9CIm3eYArNsJ0j
	XlM8mw3LkIthABEBAAGInwQYAQIACQUCUsdthgIbDAAKCRBVSiCHf5i7zrpRA/9r
	lIf6ozk+OvF6Cul7fN+8OOSUD6S6ohh/SiYKha1MSTMNWyBNhutOjmOoQoHhPmAv
	Kp8tvYULV4SiKrlCP9ANait2gmYcKsqk/kI7xel4tIvx64EMAsgaKWN7hp3TG77Y
	cVNCjtHerHjGZbRw6/GGlNSbw8DRQ0FbsPkasuexEw==
	=jr2t
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
	pvk, err := GetPrivateKeyFromKeyring(pvkr, "dyst0ni3@gmail.com")
	if err != nil {
		t.Fatalf("Unable to extract private key: %d.\n", err.Error())
	}

	pubk, err := GetPublicKeysFromKeyring([]byte(publicKey))
	if err != nil {
		t.Fatalf("Unable to extract keyring from file: %s.\n", err.Error())
	}

	plainbytes := []byte(plaintex)
	_, err = OpenPgpEncrypt(plainbytes, pvk, pubk)
	if err != nil {
		t.Fatalf("Unexpected error: %s.\n", err.Error())
	}
}
