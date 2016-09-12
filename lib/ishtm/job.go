//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

// Golang std packages
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
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
	GlobalEncryptionKey  []byte
	GlobalEncryptionSalt []byte
)

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

	tokenBytes, err := hotp.Marshal(token)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal hotp token, cause %s", err.Error())
	}

	// encrypt seed
	tokenEnc, err := crypto3n4.AesEncrypt(
		GlobalEncryptionKey,
		GlobalEncryptionSalt,
		tokenBytes,
		crypto3n4.CBC,
	)
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
		EncryptedAuthKey: authKeyEnc,
		EncryptedSeed:    tokenEnc,
	}, qr, nil
}

// generateJobID generate an hashed string id from user and data
// dependand values randomised with random bytes.
func generateJobID(owner *OwnerID, reference []byte, actual *time.Time) (string, error) {
	buf := make([]byte, 0)
	buf = append(buf, []byte(fmt.Sprintf("%s%d%d", owner.Name, actual.Unix(), actual.UnixNano()))...)
	referenceHash := sha256.Sum256(reference)
	buf = append(buf, referenceHash[:]...)
	random, err := ct.RandomBytesForLen(32)
	if err != nil {
		return "", err
	}
	buf = append(buf, random...)
	// hash generated blob
	checksum := sha256.Sum256(buf)

	return hex.EncodeToString(checksum[:]), nil
}

// NewJob init a new job struct with argument passed parameters
// validating to avoid nil values. Returns a new Job instance, a
// QR code to seed OTP generation and an error if something went
// wrong.
func NewJob(owner *OwnerID, reference []byte, settings *Settings, recipients []Recipient) (*Job, []byte, error) {
	if owner == nil ||
		settings == nil {
		return nil, nil, fmt.Errorf("provided arguments should not be nil")
	}
	if len(reference) == 0 {
		return nil, nil, fmt.Errorf("invalid reference file should not be empty")
	}

	now := time.Now().UTC()
	id, err := generateJobID(owner, reference, &now)
	if err != nil {
		return nil, nil, err
	}

	if settings.ExtensionUnit <= 0 ||
		settings.DeliveryOffset <= 0 {
		return nil, nil, fmt.Errorf("invalid offset, should never be zero or negative")
	}
	// defer ttd
	ttd := now.Add(settings.ExtensionUnit)
	if settings.DisableOffset != true {
		ttd = ttd.Add(settings.DeliveryOffset)
	}

	// generate basic auth methods
	basicCredential, qrcode, err := generateCredential()
	if err != nil {
		return nil, nil, err
	}
	owner.Credentials = []Credential{*basicCredential}

	// init structure
	job := &Job{
		ID:             id,
		Owner:          *owner,
		ReferenceFile:  reference,
		Recipients:     recipients,
		Creation:       now,
		LastModified:   now,
		LastPing:       now,
		TimeToDelivery: ttd,
		Settings:       *settings,
	}
	return job, qrcode, nil
}

// Refresh reference time to delivery deadline.
func (j *Job) Refresh() error {
	now := time.Now().UTC()
	if j.Settings.ExtensionUnit <= 0 ||
		j.Settings.DeliveryOffset <= 0 {
		return fmt.Errorf("invalid offset, should never be zero or negative")
	}
	// defer ttd
	ttd := j.TimeToDelivery.Add(j.Settings.ExtensionUnit)
	if j.Settings.DisableOffset != true {
		ttd = ttd.Add(j.Settings.DeliveryOffset)
	}
	j.TimeToDelivery = ttd

	// update time stamps
	j.LastPing = now

	return nil
}
