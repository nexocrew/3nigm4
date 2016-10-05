//
// 3nigm4 ishtmdb package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtmdb

// Golang std libs
import (
	"os"
	"time"
)

// Internal packages
import (
	types "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
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

// Mongodb database, wrapping mgo session
// structure.
type Mongodb struct {
	session *mgo.Session
	// target nodes
	database       string
	jobsCollection string
}

// MgoSession get a new session starting from the standard args
// structure.
func MgoSession(args *types.DbArgs) (*Mongodb, error) {
	s, err := mgo.Dial(types.ComposeDbAddress(args))
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
func (d *Mongodb) Copy() types.Database {
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
func (d *Mongodb) GetWills(owner string) ([]will.Will, error) {
	// build query
	selector := bson.M{
		"owner.name": bson.M{"$eq": owner},
	}
	// perform db query
	var wills []will.Will
	err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).All(&wills)
	if err != nil {
		return nil, err
	}
	return wills, nil
}

// GetWill get will structure from a given jobID, if
// something wrong returns an error.
func (d *Mongodb) GetWill(id string) (*will.Will, error) {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": id},
	}
	// perform db query
	var will will.Will
	err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).One(&will)
	if err != nil {
		return nil, err
	}
	return &will, nil
}

// SetWill upsert an argument Will struct to the database,
// returns an error if something went wrong.
func (d *Mongodb) SetWill(will *will.Will) error {
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
func (d *Mongodb) GetInDelivery(actual time.Time) ([]will.Will, error) {
	// build query
	selector := bson.M{
		"ttd": bson.M{
			"$lt": actual.UTC(),
		},
	}
	// perform db query
	var wills []will.Will
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