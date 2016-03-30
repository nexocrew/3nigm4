//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//
package message

import (
	"testing"
)

const (
	kCreatorId = "user.test@mail.com"
	kPreshared = "presharedsec"
)

func TestEncryptMessage(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared))
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}

}
