//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

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
	databaseName          = "ishtm"
	jobsCollectionName    = "jobs"
	envDatabaseName       = "NEXO_ISHTM_DATABASE"
	envJobsCollectionName = "NEXO_ISHTM_USERS_COLLECTION"
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
	// job behaviour
	GetWills(string) ([]Will, error) // list wills for owner's username.
	GetWill(string) (*Will, error)   // gets a will struct from an argument jobID;
	SetWill(*Will) error             // upsert a will in the db;
	// ttd behaviour
	GetInDelivery(time.Time) ([]Will, error)
}

// Mongodb database, wrapping mgo session
// structure.
type Mongodb struct {
	session *mgo.Session
	// target nodes
	database       string
	jobsCollection string
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
	env := os.Getenv(envDatabaseName)
	if env != "" {
		db.database = env
	} else {
		db.database = databaseName
	}
	env = os.Getenv(envJobsCollectionName)
	if env != "" {
		db.jobsCollection = env
	} else {
		db.jobsCollection = jobsCollectionName
	}
	// connect to db
	return db, nil
}

// Copy the internal session to permitt multi corutine usage.
func (d *Mongodb) Copy() Database {
	return &Mongodb{
		session:        d.session.Copy(),
		database:       d.database,
		jobsCollection: d.jobsCollection,
	}
}

// Close releases the session client.
func (d *Mongodb) Close() {
	d.session.Close()
}

// GetWills retrieve all wills related to a specified user.
func (d *Mongodb) GetWills(owner string) ([]Will, error) {
	// build query
	selector := bson.M{
		"owner.name": bson.M{"$eq": owner},
	}
	// perform db query
	var wills []Will
	err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).All(&wills)
	if err != nil {
		return nil, err
	}
	return wills, nil
}

// GetWill get will structure from a given jobID, if
// something wrong returns an error.
func (d *Mongodb) GetWill(id string) (*Will, error) {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": id},
	}
	// perform db query
	var will Will
	err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).One(&will)
	if err != nil {
		return nil, err
	}
	return &will, nil
}

// SetWill upsert an argument Will struct to the database,
// returns an error if something went wrong.
func (d *Mongodb) SetWill(will *Will) error {
	selector := bson.M{
		"id": will.ID,
	}
	update := bson.M{
		"$set": will,
	}
	_, err := d.session.DB(d.database).C(d.jobsCollection).Upsert(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveWill remove an existing will from the db.
func (d *Mongodb) RemoveWill(id string) error {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": id},
	}
	// perform db remove
	err := d.session.DB(d.database).C(d.jobsCollection).Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// GetInDelivery returns wills having passed by the actual
// time stamp.
func (d *Mongodb) GetInDelivery(actual time.Time) ([]Will, error) {
	// build query
	selector := bson.M{
		"ttd": bson.M{
			"$lt": actual.UTC(),
		},
	}
	// perform db query
	var wills []Will
	err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).All(&wills)
	if err != nil {
		return nil, err
	}
	return wills, nil
}

// EnsureMongodbIndexes assign mongodb indexes to the right
// collections, this should be done only the first time the
// collection is created.
func (d *Mongodb) EnsureMongodbIndexes() error {
	willIndex := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}
	ttdIndex := mgo.Index{
		Key:        []string{"ttd"},
		Unique:     false,
		Background: true,
		Sparse:     false,
	}
	ownerIndex := mgo.Index{
		Key:        []string{"owner.name"},
		Unique:     false,
		Background: true,
		Sparse:     false,
	}
	err := d.session.DB(d.database).C(d.jobsCollection).EnsureIndex(willIndex)
	if err != nil {
		return err
	}
	err = d.session.DB(d.database).C(d.jobsCollection).EnsureIndex(ttdIndex)
	if err != nil {
		return err
	}
	err = d.session.DB(d.database).C(d.jobsCollection).EnsureIndex(ownerIndex)
	if err != nil {
		return err
	}
	return nil
}
