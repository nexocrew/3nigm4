//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// Database wrapper used to permitt, defining a db
// interface, avoiding integration tests using a real
// mongodb instance. A mockdb struct is defined in the
// database_test.go file to implement offline tests.
// In production this file is a simple wrapper around
// mgo package.
//
package main

// Golang std libs
import (
	"fmt"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
)

// dbArgs is the exposed arguments
// required by each database interface
// implementing structs.
type dbArgs struct {
	addresses []string // cluster addresses in form <addr>:<port>;
	user      string   // authentication username;
	password  string   // authentication password;
	authDb    string   // the auth db.
}

// database an interface defyining a generic
// db, package targeting, implementation.
type database interface {
	// db client related functions
	Copy() database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
	// user behaviour
	GetUser(string) (*User, error) // gets a user struct from an argument username;
	SetUser(*User) error           // creates a new user in the db;
	// session behaviour
	GetSession([]byte) (*Session, error) // search for a session in the db;
	SetSession(*Session) error           // insert a session in the db.
}

// mongodb database, wrapping mgo session
// structure.
type mongodb struct {
	session *mgo.Session
}

// composeDbAddress compose a string starting from dbArgs slice.
func composeDbAddress(args *dbArgs) string {
	dbAccess := fmt.Sprintf("mongodb://%s:%s@", args.user, args.password)
	for idx, addr := range args.addresses {
		dbAccess += addr
		if idx != len(args.addresses)-1 {
			dbAccess += ","
		}
	}
	dbAccess += fmt.Sprintf("/?authSource=%s", args.authDb)
	return dbAccess
}

// mgoSession get a new session starting from the standard args
// structure.
func mgoSession(args *dbArgs) (*mongodb, error) {
	s, err := mgo.Dial(composeDbAddress(args))
	if err != nil {
		return nil, err
	}
	// connect to db
	return &mongodb{
		session: s,
	}, nil
}

// Copy the internal session to permitt multi corutine usage.
func (d *mongodb) Copy() database {
	return &mongodb{
		session: d.session.Copy(),
	}
}

// Close releases the session client.
func (d *mongodb) Close() {
	d.session.Close()
}

// GetUser get user strucutre from a given username, if
// something wrong returns an error.
func (d *mongodb) GetUser(username string) (*User, error) {
	return nil, nil
}

// SetUser adds an argument User struct to the database,
// returns an error if something went wrong.
func (d *mongodb) SetUser(user *User) error {
	return nil
}

// GetSession check if a session is available and still valid
// veryfing time of last seen contact against pre-defined
// timeout value.
func (d *mongodb) GetSession(token []byte) (*Session, error) {
	return nil, nil
}

// SetSession add a session data to the database.
func (d *mongodb) SetSession(*Session) error {
	return nil
}
