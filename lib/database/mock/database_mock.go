//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// This mock database is used for tests purposes, should
// never be used in production environment. It's not
// concurrency safe and do not implement any performance
// optimisation logic.
//

package dbmock

// Golang std libs
import (
	"encoding/hex"
	"fmt"
)

import (
	aty "github.com/nexocrew/3nigm4/lib/auth/types"
	dty "github.com/nexocrew/3nigm4/lib/database/types"
)

type Mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	userStorage    map[string]*aty.User
	sessionStorage map[string]*aty.Session
	// in memory storage
	fileLogStorage map[string]*dty.FileLog
	asyncTxStorage map[string]*dty.AsyncTx
}

func NewMockDb(args *dty.DbArgs) *Mockdb {
	return &Mockdb{
		addresses:      composeDbAddress(args),
		user:           args.User,
		password:       args.Password,
		authDb:         args.AuthDb,
		userStorage:    make(map[string]*aty.User),
		sessionStorage: make(map[string]*aty.Session),
		fileLogStorage: make(map[string]*dty.FileLog),
		asyncTxStorage: make(map[string]*dty.AsyncTx),
	}
}

func (d *Mockdb) Copy() dty.Database {
	return d
}

func (d *Mockdb) Close() {
}

func (d *Mockdb) GetUser(username string) (*aty.User, error) {
	user, ok := d.userStorage[username]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s user", username)
	}
	return user, nil
}

func (d *Mockdb) SetUser(user *aty.User) error {
	_, ok := d.userStorage[user.Username]
	if ok {
		return fmt.Errorf("user %s already exist in the db", user.Username)
	}
	d.userStorage[user.Username] = user
	return nil
}

func (d *Mockdb) RemoveUser(username string) error {
	if _, ok := d.userStorage[username]; !ok {
		return fmt.Errorf("unable to find required %s user", username)
	}
	delete(d.userStorage, username)
	return nil
}

func (d *Mockdb) GetSession(token []byte) (*aty.Session, error) {
	h := hex.EncodeToString(token)
	session, ok := d.sessionStorage[h]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s session", h)
	}
	return session, nil
}

func (d *Mockdb) SetSession(s *aty.Session) error {
	h := hex.EncodeToString(s.Token)
	d.sessionStorage[h] = s
	return nil
}

func (d *Mockdb) RemoveSession(token []byte) error {
	h := hex.EncodeToString(token)
	if _, ok := d.sessionStorage[h]; !ok {
		return fmt.Errorf("unable to find required %s session", h)
	}
	delete(d.sessionStorage, h)
	return nil
}

func (d *Mockdb) RemoveAllSessions() error {
	d.sessionStorage = make(map[string]*aty.Session)
	return nil
}

func (d *Mockdb) GetFileLog(filename string) (*dty.FileLog, error) {
	fl, ok := d.fileLogStorage[filename]
	if !ok {
		return nil, fmt.Errorf("unable to find a log for the required %s file", filename)
	}
	return fl, nil
}

func (d *Mockdb) SetFileLog(fl *dty.FileLog) error {
	_, ok := d.fileLogStorage[fl.Id]
	if ok {
		return fmt.Errorf("file %s already exist in the db", fl.Id)
	}
	d.fileLogStorage[fl.Id] = fl
	return nil
}

func (d *Mockdb) UpdateFileLog(fl *dty.FileLog) error {
	_, ok := d.fileLogStorage[fl.Id]
	if !ok {
		return fmt.Errorf("file %s do not exist in the db", fl.Id)
	}
	d.fileLogStorage[fl.Id] = fl
	return nil
}

func (d *Mockdb) RemoveFileLog(filename string) error {
	delete(d.fileLogStorage, filename)
	return nil
}

func (d *Mockdb) GetAsyncTx(id string) (*dty.AsyncTx, error) {
	at, ok := d.asyncTxStorage[id]
	if !ok {
		return nil, fmt.Errorf("unable to find an async tx for the required %s id", id)
	}
	return at, nil
}

func (d *Mockdb) SetAsyncTx(at *dty.AsyncTx) error {
	_, ok := d.asyncTxStorage[at.Id]
	if ok {
		return fmt.Errorf("tx %s already exist in the db", at.Id)
	}
	d.asyncTxStorage[at.Id] = at
	return nil
}

func (d *Mockdb) UpdateAsyncTx(at *dty.AsyncTx) error {
	_, ok := d.asyncTxStorage[at.Id]
	if !ok {
		return fmt.Errorf("tx %s do not exist in the db", at.Id)
	}
	d.asyncTxStorage[at.Id] = at
	return nil
}

func (d *Mockdb) RemoveAsyncTx(id string) error {
	delete(d.asyncTxStorage, id)
	return nil
}

// composeDbAddress compose a string starting from dbArgs slice.
func composeDbAddress(args *dty.DbArgs) string {
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
