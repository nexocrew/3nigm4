//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

// Package auth is a database wrapper used to permitt,
// defining a db interface, avoiding integration tests
// using a real mongodb instance. A mockdb struct is
// defined in the database_test.go file to implement
// offline tests.
// In production this file is a simple wrapper around
// mgo package.
package auth

// Golang std libs
import (
	"fmt"
	"os"
	"time"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	kDatabaseName              = "authentication"
	kUsersCollectionName       = "users"
	kSessionsCollectionName    = "session"
	kEnvDatabaseName           = "NEXO_AUTH_DATABASE"
	kEnvUsersCollectionName    = "NEXO_AUTH_USERS_COLLECTION"
	kEnvSessionsCollectionName = "NEXO_AUTH_SESSIONS_COLLECTION"
	kMaxSessionExistance       = 24 * time.Hour
)

// DbArgs is the exposed arguments
// required by each database interface
// implementing structs.
type DbArgs struct {
	Addresses []string // cluster addresses in form <addr>:<port>;
	User      string   // authentication username;
	Password  string   // authentication password;
	AuthDb    string   // the auth db.
}

// Database an interface defining a generic
// db, package targeting, implementation.
type Database interface {
	// db client related functions
	Copy() Database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
	// user behaviour
	GetUser(string) (*User, error) // gets a user struct from an argument username;
	SetUser(*User) error           // creates a new user in the db;
	RemoveUser(string) error       // remove an user from the db;
	// session behaviour
	GetSession([]byte) (*Session, error) // search for a session in the db;
	SetSession(*Session) error           // insert a session in the db;
	RemoveSession([]byte) error          // remove an existing session;
	RemoveAllSessions() error            // remove all sessions in the db.
}

// Mongodb database, wrapping mgo session
// structure.
type Mongodb struct {
	session *mgo.Session
	// target nodes
	database           string
	usersCollection    string
	sessionsCollection string
}

// composeDbAddress compose a string starting from dbArgs slice.
func composeDbAddress(args *DbArgs) string {
	dbAccess := fmt.Sprintf("mongodb://%s:%s@", args.User, args.Password)
	for idx, addr := range args.Addresses {
		dbAccess += addr
		if idx != len(args.Addresses)-1 {
			dbAccess += ","
		}
	}
	dbAccess += fmt.Sprintf("/?authSource=%s", args.AuthDb)
	return dbAccess
}

// MgoSession get a new session starting from the standard args
// structure.
func MgoSession(args *DbArgs) (*Mongodb, error) {
	s, err := mgo.Dial(composeDbAddress(args))
	if err != nil {
		return nil, err
	}
	db := &Mongodb{
		session: s,
	}
	// check for env vars
	env := os.Getenv(kEnvDatabaseName)
	if env != "" {
		db.database = env
	} else {
		db.database = kDatabaseName
	}
	env = os.Getenv(kEnvSessionsCollectionName)
	if env != "" {
		db.sessionsCollection = env
	} else {
		db.sessionsCollection = kSessionsCollectionName
	}
	env = os.Getenv(kEnvUsersCollectionName)
	if env != "" {
		db.usersCollection = env
	} else {
		db.usersCollection = kUsersCollectionName
	}
	// connect to db
	return db, nil
}

// Copy the internal session to permitt multi corutine usage.
func (d *Mongodb) Copy() Database {
	return &Mongodb{
		session:            d.session.Copy(),
		database:           d.database,
		usersCollection:    d.usersCollection,
		sessionsCollection: d.sessionsCollection,
	}
}

// Close releases the session client.
func (d *Mongodb) Close() {
	d.session.Close()
}

// GetUser get user strucutre from a given username, if
// something wrong returns an error.
func (d *Mongodb) GetUser(username string) (*User, error) {
	// build query
	selector := bson.M{
		"username": bson.M{"$eq": username},
	}
	// perform db query
	var user User
	err := d.session.DB(d.database).C(d.usersCollection).Find(selector).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUser adds an argument User struct to the database,
// returns an error if something went wrong.
func (d *Mongodb) SetUser(user *User) error {
	selector := bson.M{
		"username": user.Username,
	}
	update := bson.M{
		"$set": user,
	}
	_, err := d.session.DB(d.database).C(d.usersCollection).Upsert(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveUser remove an existing user from the db.
func (d *Mongodb) RemoveUser(username string) error {
	// build query
	selector := bson.M{
		"username": bson.M{"$eq": username},
	}
	// perform db remove
	err := d.session.DB(d.database).C(d.usersCollection).Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// GetSession check if a session is available and still valid
// veryfing time of last seen contact against pre-defined
// timeout value.
func (d *Mongodb) GetSession(token []byte) (*Session, error) {
	// build query
	selector := bson.M{
		"token": bson.M{"$eq": token},
	}
	// perform db query
	var session Session
	err := d.session.DB(d.database).C(d.sessionsCollection).Find(selector).One(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// SetSession add a session data to the database.
func (d *Mongodb) SetSession(session *Session) error {
	selector := bson.M{
		"token": session.Token,
	}
	update := bson.M{
		"$set": session,
	}
	_, err := d.session.DB(d.database).C(d.sessionsCollection).Upsert(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveSession remove a session from the db.
func (d *Mongodb) RemoveSession(token []byte) error {
	// build query
	selector := bson.M{
		"token": bson.M{"$eq": token},
	}
	// perform db remove
	err := d.session.DB(d.database).C(d.sessionsCollection).Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// RemoveAllSessions remove all active and not active
// sessions from the database instance.
func (d *Mongodb) RemoveAllSessions() error {
	// perform db remove all
	_, err := d.session.DB(d.database).C(d.sessionsCollection).RemoveAll(bson.M{})
	if err != nil {
		return err
	}
	return nil
}

// EnsureMongodbIndexes assign mongodb indexes to the right
// collections, this should be done only the first time the
// collection is created.
func (d *Mongodb) EnsureMongodbIndexes() error {
	usersIndex := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}
	err := d.session.DB(d.database).C(d.usersCollection).EnsureIndex(usersIndex)
	if err != nil {
		return err
	}
	sessionIndex := mgo.Index{
		Key:        []string{"token"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}
	userSessionIndex := mgo.Index{
		Key:        []string{"username"},
		Unique:     false,
		Background: true,
		Sparse:     false,
	}
	// the following index is used to
	// clean out every session after 32 hours
	// from the creation time.
	cleanSessionIndex := mgo.Index{
		Key:         []string{"lastseen_ts"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: kMaxSessionExistance, // clean session at max every 24 hours (time.Duration type).
	}
	err = d.session.DB(d.database).C(d.sessionsCollection).EnsureIndex(sessionIndex)
	if err != nil {
		return err
	}
	err = d.session.DB(d.database).C(d.sessionsCollection).EnsureIndex(userSessionIndex)
	if err != nil {
		return err
	}
	err = d.session.DB(d.database).C(d.sessionsCollection).EnsureIndex(cleanSessionIndex)
	if err != nil {
		return err
	}
	return nil
}
