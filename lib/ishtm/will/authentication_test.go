//
// 3nigm4 will package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package will

// Golang std packages
import (
	"encoding/hex"
	"fmt"
	"testing"
)

// Third party packages
import (
	"github.com/gokyle/hotp"
)

func getSwToken(credentials *Credential) (*hotp.HOTP, error) {
	if credentials == nil {
		return nil, fmt.Errorf("argument credentials is required and should not be nil")
	}

	// decrypt token content
	swtoken, err := decryptHotp(credentials.SoftwareToken)
	if err != nil {
		return nil, err
	}

	return swtoken, nil
}

func TestOTPVerification(t *testing.T) {
	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	creds, _, _, err := generateCredential()
	if err != nil {
		t.Fatalf("Unable to generate credentials: %s.\n", err.Error())
	}
	//  get swtoken
	clientToken, err := getSwToken(creds)
	if err != nil {
		t.Fatalf("Software token extraction error: %s.\n", err.Error())
	}
	otp := clientToken.OTP()
	t.Logf("OTP: %s.\n", otp)

	// verify otp
	verificationCreds, err := verifyOTP(otp, creds)
	if err != nil {
		t.Fatalf("Verification failed: %s.\n", err.Error())
	}

	otp = clientToken.OTP()
	t.Logf("OTP: %s.\n", otp)

	verificationCreds, err = verifyOTP(otp, verificationCreds)
	if err != nil {
		t.Fatalf("Verification failed: %s.\n", err.Error())
	}

	for i := 0; i < ckeckIncrementTolerance; i++ {
		otp = clientToken.OTP()
		t.Logf("OTP: %s.\n", otp)
	}

	verificationCreds, err = verifyOTP(otp, verificationCreds)
	if err != nil {
		t.Fatalf("Verification failed: %s.\n", err.Error())
	}
}

func TestOTPOutRange(t *testing.T) {
	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	creds, _, _, err := generateCredential()
	if err != nil {
		t.Fatalf("Unable to generate credentials: %s.\n", err.Error())
	}
	//  get swtoken
	clientToken, err := getSwToken(creds)
	if err != nil {
		t.Fatalf("Software token extraction error: %s.\n", err.Error())
	}
	otp := clientToken.OTP()
	t.Logf("OTP: %s.\n", otp)

	// verify otp
	verificationCreds, err := verifyOTP(otp, creds)
	if err != nil {
		t.Fatalf("Verification failed: %s.\n", err.Error())
	}

	for i := 0; i < ckeckIncrementTolerance+1; i++ {
		otp = clientToken.OTP()
		t.Logf("OTP: %s.\n", otp)
	}

	verificationCreds, err = verifyOTP(otp, verificationCreds)
	if err == nil {
		t.Fatalf("Verification must fail if async is more than %d clicks.\n", ckeckIncrementTolerance)
	}
}

func TestSecondaryKeys(t *testing.T) {
	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	creds, _, keys, err := generateCredential()
	if err != nil {
		t.Fatalf("Unable to generate credentials: %s.\n", err.Error())
	}

	if len(keys) != secondaryKeysNumber {
		t.Fatalf("Unexpected number of secondary keys, having %d expecting %d.\n", len(keys), secondaryKeysNumber)
	}

	keyRaw, err := hex.DecodeString(keys[0])
	if err != nil {
		t.Fatalf("Unable to decode hex: %s.\n", err.Error())
	}

	modifyiedCreds, err := verifySecondaryKeys(keyRaw, creds)
	if err != nil {
		t.Fatalf("Unrecognised key: %s.\n", err.Error())
	}
	if len(modifyiedCreds.SecondaryKeys) != secondaryKeysNumber-1 {
		t.Fatalf("Unexpected number of secondary keys, having %d expecting %d.\n", len(modifyiedCreds.SecondaryKeys), secondaryKeysNumber-1)
	}

	keyRaw, err = hex.DecodeString(keys[1])
	if err != nil {
		t.Fatalf("Unable to decode hex: %s.\n", err.Error())
	}

	modifyiedCreds, err = verifySecondaryKeys(keyRaw, creds)
	if err != nil {
		t.Fatalf("Unrecognised key: %s.\n", err.Error())
	}
	if len(modifyiedCreds.SecondaryKeys) != secondaryKeysNumber-2 {
		t.Fatalf("Unexpected number of secondary keys, having %d expecting %d.\n", len(modifyiedCreds.SecondaryKeys), secondaryKeysNumber-2)
	}

	keyRaw, err = hex.DecodeString(keys[2])
	if err != nil {
		t.Fatalf("Unable to decode hex: %s.\n", err.Error())
	}

	modifyiedCreds, err = verifySecondaryKeys(keyRaw, creds)
	if err != nil {
		t.Fatalf("Unrecognised key: %s.\n", err.Error())
	}
	if len(modifyiedCreds.SecondaryKeys) != secondaryKeysNumber-3 {
		t.Fatalf("Unexpected number of secondary keys, having %d expecting %d.\n", len(modifyiedCreds.SecondaryKeys), secondaryKeysNumber-3)
	}

	keyRaw, err = hex.DecodeString(keys[1])
	if err != nil {
		t.Fatalf("Unable to decode hex: %s.\n", err.Error())
	}

	modifyiedCreds, err = verifySecondaryKeys(keyRaw, creds)
	if err == nil {
		t.Fatalf("An already used key should not be usable again.\n")
	}
}
