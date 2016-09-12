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
)

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

func NewJob(owner *OwnerID, reference []byte, settings *Settings, recipients []Recipient) (*Job, error) {
	now := time.Now().UTC()
	id, err := generateJobID(owner, reference, &now)
	if err != nil {
		return nil, err
	}

	// define ttd
	ttd := now.Add(settings.ExtensionUnit)
	if settings.DisableOffset != true {
		tdd = ttd.Add(settings.DeliveryOffset)
	}

	// init structure
	job := &Job{
		ID:             id,
		Owner:          *OwnerID,
		ReferenceFile:  reference,
		Recipients:     recipients,
		Creation:       now,
		LastModified:   now,
		LastPing:       now,
		TimeToDelivery: ttd,
		Settings:       *Settings,
	}
	return job, nil
}
