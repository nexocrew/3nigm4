package types

// Golang std lib.
import (
	"time"
)

import (
	aty "github.com/nexocrew/3nigm4/lib/auth/types"
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Database an interface defining a generic
// db, package targeting, implementation.
type Database interface {
	// db client related functions
	Copy() Database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
	// user behaviour
	GetUser(string) (*aty.User, error) // gets a user struct from an argument username;
	SetUser(*aty.User) error           // creates a new user in the db;
	RemoveUser(string) error           // remove an user from the db;
	// session behaviour
	GetSession([]byte) (*aty.Session, error) // search for a session in the db;
	SetSession(*aty.Session) error           // insert a session in the db;
	RemoveSession([]byte) error              // remove an existing session;
	RemoveAllSessions() error                // remove all sessions in the db.
	// db create file log
	SetFileLog(fl *FileLog) error             // add a new file log when a file is uploaded;
	UpdateFileLog(fl *FileLog) error          // update an existing file log;
	GetFileLog(file string) (*FileLog, error) // get infos to a previously uploaded file;
	RemoveFileLog(file string) error          // remove a previously added file log;
	// async tx
	SetAsyncTx(at *AsyncTx) error           // add a new async tx record;
	UpdateAsyncTx(at *AsyncTx) error        // update an existing async tx;
	GetAsyncTx(id string) (*AsyncTx, error) // get an existing tx;
	RemoveAsyncTx(id string) error          // remove an existing tx (typically should be done automatically with a ttl setup).
}

// DbArgs is the exposed arguments
// required by each database interface
// implementing structs.
type DbArgs struct {
	Addresses []string // cluster addresses in form <addr>:<port>;
	User      string   // authentication username;
	Password  string   // authentication password;
	AuthDb    string   // the auth db.
}

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
