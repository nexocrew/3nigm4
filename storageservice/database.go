//
// 3nigm4 storageservice package
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
	"os"
	"time"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	defaultDatabaseName           = "storageservice"
	defaultFilesLogCollectionName = "fileslog"
	defaultAsyncTxCollectionName  = "asynctx"
	envDatabaseName               = "NEXO_FILESLOG_DATABASE"
	envFilesLogCollectionName     = "NEXO_FILESLOG_COLLECTION"
	envAsyncTxCollectionName      = "NEXO_ASYNCTX_COLLECTION"
)

// MaxAsyncTxExistance represent the maximum time
// that an async job can remain pending in the database
// before being automatically deleted.
var MaxAsyncTxExistance = 1 * time.Hour

// dbArgs is the exposed arguments
// required by each database interface
// implementing structs.
type dbArgs struct {
	addresses []string // cluster addresses in form <addr>:<port>;
	user      string   // authentication username;
	password  string   // authentication password;
	authDb    string   // the auth db.
}

// database an interface defining a generic
// db, package targeting, implementation.
type database interface {
	// db client related functions
	Copy() database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
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

// mongodb database, wrapping mgo session
// structure.
type mongodb struct {
	session *mgo.Session
	// target nodes
	databaseName      string
	filelogCollection string
	asyncTxCollection string
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

// MgoSession get a new session starting from the standard args
// structure.
func MgoSession(args *dbArgs) (*mongodb, error) {
	s, err := mgo.Dial(composeDbAddress(args))
	if err != nil {
		return nil, err
	}
	db := &mongodb{
		session: s,
	}
	// check for env vars
	env := os.Getenv(envDatabaseName)
	if env != "" {
		db.databaseName = env
	} else {
		db.databaseName = defaultDatabaseName
	}
	env = os.Getenv(envFilesLogCollectionName)
	if env != "" {
		db.filelogCollection = env
	} else {
		db.filelogCollection = defaultFilesLogCollectionName
	}
	env = os.Getenv(envAsyncTxCollectionName)
	if env != "" {
		db.asyncTxCollection = env
	} else {
		db.asyncTxCollection = defaultAsyncTxCollectionName
	}
	// connect to db
	return db, nil
}

// Copy the internal session to permitt multi corutine usage.
func (d *mongodb) Copy() database {
	return &mongodb{
		session:           d.session.Copy(),
		databaseName:      d.databaseName,
		filelogCollection: d.filelogCollection,
		asyncTxCollection: d.asyncTxCollection,
	}
}

// Close releases the session client.
func (d *mongodb) Close() {
	d.session.Close()
}

// GetFileLog get file informations from the db.
func (d *mongodb) GetFileLog(filename string) (*FileLog, error) {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": filename},
	}
	// perform db query
	var filelog FileLog
	err := d.session.DB(d.databaseName).C(d.filelogCollection).Find(selector).One(&filelog)
	if err != nil {
		return nil, err
	}
	return &filelog, nil
}

// SetFileLog add a new file log to the database
// this operation will tipically be performed after
// uploading successfully a data file to the S3
// backend.
func (d *mongodb) SetFileLog(fl *FileLog) error {
	err := d.session.DB(d.databaseName).C(d.filelogCollection).Insert(fl)
	if err != nil {
		return err
	}
	return nil
}

// UpdateFileLog update a previously created document
// with updated argument structure.
func (d *mongodb) UpdateFileLog(fl *FileLog) error {
	selector := bson.M{
		"id": fl.Id,
	}
	update := bson.M{
		"$set": fl,
	}
	err := d.session.DB(d.databaseName).C(d.filelogCollection).Update(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveFileLog remove an existing file log from the db.
func (d *mongodb) RemoveFileLog(filename string) error {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": filename},
	}
	// perform db remove
	err := d.session.DB(d.databaseName).C(d.filelogCollection).Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// GetAsyncTx returns an async tx document from the mongodb
// instance.
func (d *mongodb) GetAsyncTx(id string) (*AsyncTx, error) {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": id},
	}
	// perform db query
	var tx AsyncTx
	err := d.session.DB(d.databaseName).C(d.asyncTxCollection).Find(selector).One(&tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// SetAsyncTx add a new async tx document to the mongodb
// instance.
func (d *mongodb) SetAsyncTx(at *AsyncTx) error {
	err := d.session.DB(d.databaseName).C(d.asyncTxCollection).Insert(at)
	if err != nil {
		return err
	}
	return nil
}

// UpdateAsyncTx update an existing tx with the argument passed
// doc.
func (d *mongodb) UpdateAsyncTx(at *AsyncTx) error {
	selector := bson.M{
		"id": at.Id,
	}
	update := bson.M{
		"$set": at,
	}
	err := d.session.DB(d.databaseName).C(d.asyncTxCollection).Update(selector, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveAsyncTx removes an existing async tx from the mongodb
// instance.
func (d *mongodb) RemoveAsyncTx(id string) error {
	// build query
	selector := bson.M{
		"id": bson.M{"$eq": id},
	}
	// perform db remove
	err := d.session.DB(d.databaseName).C(d.asyncTxCollection).Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// ensureMongodbIndexes assign mongodb indexes to the right
// collections, this should be done only the first time the
// collection is created.
func (d *mongodb) EnsureMongodbIndexes() error {
	// file log
	fileIndex := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}
	err := d.session.DB(d.databaseName).C(d.filelogCollection).EnsureIndex(fileIndex)
	if err != nil {
		return err
	}
	// async tx
	idIndex := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		Background: true,
		Sparse:     false,
	}
	// the following index is used to
	// clean out every async tx after 1 hours
	// from the creation time.
	ttlIndex := mgo.Index{
		Key:         []string{"ts"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: MaxAsyncTxExistance, // clean async tx at max every 1 hours (time.Duration type).
	}
	err = d.session.DB(d.databaseName).C(d.asyncTxCollection).EnsureIndex(idIndex)
	if err != nil {
		return err
	}
	err = d.session.DB(d.databaseName).C(d.asyncTxCollection).EnsureIndex(ttlIndex)
	if err != nil {
		return err
	}
	return nil
}
