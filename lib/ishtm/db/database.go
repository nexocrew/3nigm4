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
	ct "github.com/nexocrew/3nigm4/lib/commons"
	types "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	databaseName            = "ishtm"
	jobsCollectionName      = "jobs"
	emailsCollectionName    = "emails"
	envDatabaseName         = "NEXO_ISHTM_DATABASE"
	envJobsCollectionName   = "NEXO_ISHTM_USERS_COLLECTION"
	envEmailsCollectionName = "NEXO_ISHTM_EMAILS"
)

// Mongodb database, wrapping mgo session
// structure.
type Mongodb struct {
	session *mgo.Session
	// target nodes
	database         string
	jobsCollection   string
	emailsCollection string
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
	env = os.Getenv(envEmailsCollectionName)
	if env != "" {
		db.emailsCollection = env
	} else {
		db.emailsCollection = emailsCollectionName
	}
	// connect to db
	return db, nil
}

// Copy the internal session to permitt multi corutine usage.
func (d *Mongodb) Copy() types.Database {
	return &Mongodb{
		session:          d.session.Copy(),
		database:         d.database,
		jobsCollection:   d.jobsCollection,
		emailsCollection: d.emailsCollection,
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

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"removable": true,
			},
		},
		ReturnNew: false,
	}
	var wills []will.Will
	_, err := d.session.DB(d.database).C(d.jobsCollection).Find(selector).Apply(change, &wills)
	if err != nil {
		return nil, err
	}
	return wills, nil
}

// RemoveExausted deletes all documents containing the "removable"
// flag setted to true
func (d *Mongodb) RemoveExausted() error {
	// build query
	selector := bson.M{
		"removable": bson.M{"$eq": true},
	}
	// perform db remove of "reovable" objects
	_, err := d.session.DB(d.database).C(d.jobsCollection).RemoveAll(selector)
	if err != nil {
		return err
	}
	return nil
}

// SetEmail upsert an email in the database to be
// sended by the dispatcher.
func (d *Mongodb) SetEmail(email *ct.Email) error {
	selector := bson.M{
		"_id": email.ObjectID,
	}
	update := bson.M{
		"$set": email,
	}
	_, err := d.session.DB(d.database).C(d.emailsCollection).Upsert(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// GetEmails returns non sended emails for providing
// the dispatcher with required emails.
func (d *Mongodb) GetEmails() ([]ct.Email, error) {
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"sended": true,
			},
		},
		ReturnNew: false,
	}
	query := bson.M{
		"sended": bson.M{
			"$eq": false,
		},
	}
	var emails []ct.Email
	_, err := d.session.DB(d.database).C(d.emailsCollection).Find(query).Apply(change, &emails)
	if err != nil {
		return nil, err
	}
	return emails, nil
}

// RemoveSendedEmails remove sended emails while possible.
func (d *Mongodb) RemoveSendedEmails() error {
	// build query
	selector := bson.M{
		"sended": bson.M{"$eq": true},
	}
	// perform db remove
	_, err := d.session.DB(d.database).C(d.emailsCollection).RemoveAll(selector)
	if err != nil {
		return err
	}
	return nil
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
