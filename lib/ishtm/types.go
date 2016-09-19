//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

// Golang std pkgs
import (
	"time"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

type Credential struct {
	SoftwareToken []byte `bson:"swtoken"`
	SecondaryKey  []byte `bson:"secondarykey"`
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
