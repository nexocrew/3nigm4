//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// This mock database is used for tests purposes, should
// never be used in production environment. It's not
// concurrency safe and do not implement any performance
// optimisation logic.
//
package main

// Golang std libs
import (
	"fmt"
)

type mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	fileLogStorage map[string]*FileLog
}

func newMockDb(args *dbArgs) *mockdb {
	return &mockdb{
		addresses:      composeDbAddress(args),
		user:           args.user,
		password:       args.password,
		authDb:         args.authDb,
		fileLogStorage: make(map[string]*FileLog),
	}
}

func (d *mockdb) Copy() database {
	return d
}

func (d *mockdb) Close() {
}

func (d *mockdb) GetFileLog(filename string) (*FileLog, error) {
	fl, ok := d.fileLogStorage[filename]
	if !ok {
		return nil, fmt.Errorf("unable to find a log for the required %s file", filename)
	}
	return fl, nil
}

func (d *mockdb) SetFileLog(fl *FileLog) error {
	_, ok := d.fileLogStorage[fl.Id]
	if ok {
		return fmt.Errorf("file %s already exist in the db", fl.Id)
	}
	d.fileLogStorage[fl.Id] = fl
	return nil
}

func (d *mockdb) RemoveFileLog(filename string) error {
	delete(d.fileLogStorage, filename)
	return nil
}
