//
// 3nigm4 ishtmtypes package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

// Package ishtmtypes defines common types and utility
// functions for the ishtm package.
package ishtmtypes

// Golang std pkgs
import (
	"time"
)

// Internal pkgs
import (
	w "github.com/nexocrew/3nigm4/lib/ishtm/will"
)

// Database an interface defining a generic
// db, package targeting, implementation.
type Database interface {
	// db client related functions
	Copy() Database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
	// job behaviour
	GetWills(string) ([]w.Will, error) // list wills for owner's username.
	GetWill(string) (*w.Will, error)   // gets a will struct from an argument jobID;
	SetWill(*w.Will) error             // upsert a will in the db;
	RemoveWill(string) error           // remove a will from the db;
	// ttd behaviour
	GetInDelivery(time.Time) ([]w.Will, error)
	RemoveExausted() error
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
