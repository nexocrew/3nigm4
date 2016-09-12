//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

// Golang std packages
import (
	"time"
)

type Recipient struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
	// pgp key identity
	KeyID       uint64 `bson:"keyid"`
	Fingerprint []byte `bson:"fingerprint"`
}

type Credential struct {
	EncryptedSeed    []byte `bson:"seed"`
	EncryptedAuthKey []byte `bson:"authkey"`
}

type OwnerID struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
	// pgp key identity
	KeyID       uint64 `bson:"keyid"`
	Fingerprint []byte `bson:"fingerprint"`
	// ping credentials
	Credentials []Credential `bson:"credentials"`
}

type Settings struct {
	ExtensionUnit  time.Duration `bson:"extensionunit"`
	NotifyDeadline bool          `bson:"notifydeadline"`
	DeliveryOffset time.Duration `bson:"deliveryoffset"`
	DisableOffset  bool          `bson:"disableoffset"`
}

type Job struct {
	ID             string      `bson:"id"`
	Owner          OwnerID     `bson:"owner"`
	Creation       time.Time   `bson:"creation"`
	LastModified   time.Time   `bson:"lastmodified"`
	ReferenceFile  []byte      `bson:"referencefile"`
	Recipients     []Recipient `bson:"recipients"`
	LastPing       time.Time   `bson:"lastping"` // UTC located;
	TimeToDelivery time.Time   `bson:"ttd"`      // UTC located;
	Settings       Settings    `bson:"settings"`
	Disabled       bool        `bson:"disabled"`
}
