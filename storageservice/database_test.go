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
	asyncTxStorage map[string]*AsyncTx
}

func newMockDb(args *dbArgs) *mockdb {
	return &mockdb{
		addresses:      composeDbAddress(args),
		user:           args.user,
		password:       args.password,
		authDb:         args.authDb,
		fileLogStorage: make(map[string]*FileLog),
		asyncTxStorage: make(map[string]*AsyncTx),
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

func (d *mockdb) UpdateFileLog(fl *FileLog) error {
	_, ok := d.fileLogStorage[fl.Id]
	if !ok {
		return fmt.Errorf("file %s do not exist in the db", fl.Id)
	}
	d.fileLogStorage[fl.Id] = fl
	return nil
}

func (d *mockdb) RemoveFileLog(filename string) error {
	delete(d.fileLogStorage, filename)
	return nil
}

func (d *mockdb) GetAsyncTx(id string) (*AsyncTx, error) {
	at, ok := d.asyncTxStorage[id]
	if !ok {
		return nil, fmt.Errorf("unable to find an async tx for the required %s id", id)
	}
	return at, nil
}

func (d *mockdb) SetAsyncTx(at *AsyncTx) error {
	_, ok := d.asyncTxStorage[at.Id]
	if ok {
		return fmt.Errorf("tx %s already exist in the db", at.Id)
	}
	d.asyncTxStorage[at.Id] = at
	return nil
}

func (d *mockdb) UpdateAsyncTx(at *AsyncTx) error {
	_, ok := d.asyncTxStorage[at.Id]
	if !ok {
		return fmt.Errorf("tx %s do not exist in the db", at.Id)
	}
	d.asyncTxStorage[at.Id] = at
	return nil
}

func (d *mockdb) RemoveAsyncTx(id string) error {
	delete(d.asyncTxStorage, id)
	return nil
}
