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
	"os"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	kDatabaseName              = "storageservice"
	kFilesLogCollectionName    = "fileslog"
	kEnvDatabaseName           = "NEXO_FILESLOG_DATABASE"
	kEnvFilesLogCollectionName = "NEXO_ILESLOG_COLLECTION"
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

// database an interface defining a generic
// db, package targeting, implementation.
type database interface {
	// db client related functions
	Copy() database // retain the db client in a multi-coroutine environment;
	Close()         // release the client;
	// db create file log
	SetFileLog(fl *FileLog) error             // add a new file log when a file is uploaded;
	GetFileLog(file string) (*FileLog, error) // get infos to a previously uploaded file;
	RemoveFileLog(file string) error          // remove a previously added file log.
}

// mongodb database, wrapping mgo session
// structure.
type mongodb struct {
	session *mgo.Session
	// target nodes
	databaseName      string
	filelogCollection string
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
func MgoSession(args *dbArgs) (*mongodb, error) {
	s, err := mgo.Dial(composeDbAddress(args))
	if err != nil {
		return nil, err
	}
	db := &mongodb{
		session: s,
	}
	// check for env vars
	env := os.Getenv(kEnvDatabaseName)
	if env != "" {
		db.databaseName = env
	} else {
		db.databaseName = kDatabaseName
	}
	env = os.Getenv(kEnvFilesLogCollectionName)
	if env != "" {
		db.filelogCollection = env
	} else {
		db.filelogCollection = kFilesLogCollectionName
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

// ensureMongodbIndexes assign mongodb indexes to the right
// collections, this should be done only the first time the
// collection is created.
func (d *mongodb) EnsureMongodbIndexes() error {
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
	return nil
}
