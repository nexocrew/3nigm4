package types

// Golang std lib.
import (
	"time"
)

import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// FileLog is used to store all informations about an uploaded file,
// this record will be used, later on, to manage access to the file and
// auto-remove policies.
type FileLog struct {
	Id         string        `bson:"id"`            // the file name assiciated with S3 saved data;
	Size       int           `bson:"size"`          // the size of the uploaded data blob;
	Bucket     string        `bson:"bucket"`        // the S3 bucket where data has been saved;
	CheckSum   ct.CheckSum   `bson:"checksum"`      // checksum for the uploaded data;
	Ownership  Owner         `bson:"ownership"`     // info related to the uploading user;
	Acl        Acl           `bson:"acl"`           // access permissions;
	Creation   time.Time     `bson:"creation_time"` // time of the upload;
	TimeToLive time.Duration `bson:"ttl,omitempty"` // time to live for the uploaded file;
	Complete   bool          `bson:"complete"`      // transaction completed.
}

// AsyncTx is the structure used to temporarly manage async
// transaction: in particular let the system manage S3 destined
// uploads that are managed via working queue and so not in sync
// with the API handlers (operations are for that reason splitted
// in two times a setting step and a verify step).
type AsyncTx struct {
	Id        string      `bson:"id"`                 // transaction id, different from resource id;
	Complete  bool        `bson:"complete"`           // transaction completed;
	Error     error       `bson:"error,omitempty"`    // error setted on if a transaction error encountered;
	Data      []byte      `bson:"data,omitempty"`     // data to be returned at the verify step;
	CheckSum  ct.CheckSum `bson:"checksum,omitempty"` // checksum for the transaction returned data, if any;
	Ownership Owner       `bson:"ownership"`          // info related to the uploading user;
	TimeStamp time.Time   `bson:"ts"`                 // transaction creation time: tx records can survice at max n mins (see db implementation).
}

// Acl describes access read rules to the file (writing on the file
// is an exclusive privilege of the uploading user).
type Acl struct {
	Permission   ct.Permission `bson:"permission"`        // the type of enforced permission;
	SharingUsers []string      `bson:"sharing,omitempty"` // usernames of users enabled to access the file (only in case of Shared permission type).
}

// Owner the file owner.
type Owner struct {
	Username  string `bson:"username"`            // the user uploading the file;
	OriginIp  string `bson:"ipaddr,omitempty"`    // origin upload ip address;
	UserAgent string `bson:"useragent,omitempty"` // client origin useragent.
}
