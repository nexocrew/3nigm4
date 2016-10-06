//
// 3nigm4 will package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package will

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

const (
	deliveryKeySize = 64 // lenght in bytes for delivery key.
)

type Credential struct {
	SoftwareToken []byte   `bson:"swtoken"`
	SecondaryKeys [][]byte `bson:"secondarykeys"`
}

type OwnerID struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
	// ping credentials
	Credentials []Credential `bson:"credentials"`
}

type Settings struct {
	ExtensionUnit  time.Duration `bson:"extensionunit"`
	NotifyDeadline bool          `bson:"notifydeadline"`
	DeliveryOffset time.Duration `bson:"deliveryoffset"`
	DisableOffset  bool          `bson:"disableoffset"`
}

type Will struct {
	ID             string         `bson:"id"`
	Owner          OwnerID        `bson:"owner"`
	Creation       time.Time      `bson:"creation"`
	LastModified   time.Time      `bson:"lastmodified"`
	ReferenceFile  []byte         `bson:"referencefile"`
	Recipients     []ct.Recipient `bson:"recipients"`
	LastPing       time.Time      `bson:"lastping"` // UTC located;
	TimeToDelivery time.Time      `bson:"ttd"`      // UTC located;
	Settings       Settings       `bson:"settings"`
	Disabled       bool           `bson:"disabled"`
	// delivery related
	DeliveryKey []byte `bson:"deliverykey"`
	Deliverable bool   `bson:"deliverable,omitempty"`
}

// NewWill init a new job struct with argument passed parameters
// validating to avoid nil values. Returns a new Will instance, a
// QR code to seed OTP generation and an error if something went
// wrong.
func NewWill(owner *OwnerID, reference []byte, settings *Settings, recipients []ct.Recipient) (*Will, *ct.WillCredentials, error) {
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
	basicCredential, qrcode, seckeys, err := generateCredential()
	if err != nil {
		return nil, nil, err
	}
	owner.Credentials = []Credential{*basicCredential}

	// generate delivery key
	deliveryKey, err := ct.RandomBytesForLen(deliveryKeySize)
	if err != nil {
		return nil, nil, err
	}

	// init structure
	will := &Will{
		ID:             id,
		Owner:          *owner,
		ReferenceFile:  reference,
		Recipients:     recipients,
		Creation:       now,
		LastModified:   now,
		LastPing:       now,
		TimeToDelivery: ttd,
		Settings:       *settings,
		DeliveryKey:    deliveryKey,
		Deliverable:    false,
	}

	return will,
		&ct.WillCredentials{
			QRCode:        qrcode,
			SecondaryKeys: seckeys,
		}, nil
}

// VerifyOtp verify otp validity and updates
// credentials structure for an argument index.
// The secondary key is used when the otp is not passed
// by the request structure.
func (j *Will) VerifyOtp(idx int, otp string, secondary string) error {
	if len(j.Owner.Credentials) <= idx {
		return fmt.Errorf("required credential index is out of bounds, havind %d but max %d", idx, len(j.Owner.Credentials))
	}
	credential := j.Owner.Credentials[idx]
	// check for available credentials
	if otp != "" {
		updated, err := verifyOTP(otp, &credential)
		if err != nil {
			return err
		}
		j.Owner.Credentials[idx] = *updated
	} else if secondary != "" {
		secondaryKey, err := hex.DecodeString(secondary)
		if err != nil {
			return fmt.Errorf("unable to decode secondary key (%s)", err.Error())
		}
		updated, err := verifySecondaryKeys(secondaryKey, &credential)
		if err != nil {
			return err
		}
		j.Owner.Credentials[idx] = *updated
	} else {
		return fmt.Errorf("unsupported nil credentials")
	}
	return nil

}

// Refresh reference time to delivery deadline.
func (j *Will) Refresh() error {
	now := time.Now().UTC()
	if j.Settings.ExtensionUnit <= 0 ||
		j.Settings.DeliveryOffset <= 0 {
		return fmt.Errorf("invalid offset, should never be zero or negative")
	}
	// defer ttd
	ttd := now.Add(j.Settings.ExtensionUnit)
	if j.Settings.DisableOffset != true {
		ttd = ttd.Add(j.Settings.DeliveryOffset)
	}
	j.TimeToDelivery = ttd

	// update time stamps
	j.LastPing = now

	return nil
}
