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
	"encoding/hex"
	"fmt"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/auth"
)

// Third party libs
import (
	"golang.org/x/crypto/bcrypt"
)

// composeDbAddress compose a string starting from dbArgs slice.
func composeDbAddress(args *auth.DbArgs) string {
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

// Default bcrypt iterations
const kBCryptIterations = 10

func bcryptPassword(pwd string) ([]byte, error) {
	pwdBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), kBCryptIterations)
	if err != nil {
		return nil, err
	}
	return pwdBytes, nil
}

type mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	userStorage    map[string]*auth.User
	sessionStorage map[string]*auth.Session
}

func newMockDb(args *auth.DbArgs) *mockdb {
	return &mockdb{
		addresses:      composeDbAddress(args),
		user:           args.User,
		password:       args.Password,
		authDb:         args.AuthDb,
		userStorage:    make(map[string]*auth.User),
		sessionStorage: make(map[string]*auth.Session),
	}
}

func (d *mockdb) Copy() auth.Database {
	return d
}

func (d *mockdb) Close() {
}

func (d *mockdb) GetUser(username string) (*auth.User, error) {
	user, ok := d.userStorage[username]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s user", username)
	}
	return user, nil
}

func (d *mockdb) SetUser(user *auth.User) error {
	_, ok := d.userStorage[user.Username]
	if ok {
		return fmt.Errorf("user %s already exist in the db", user.Username)
	}
	d.userStorage[user.Username] = user
	return nil
}

func (d *mockdb) RemoveUser(username string) error {
	if _, ok := d.userStorage[username]; !ok {
		return fmt.Errorf("unable to find required %s user", username)
	}
	delete(d.userStorage, username)
	return nil
}

func (d *mockdb) GetSession(token []byte) (*auth.Session, error) {
	h := hex.EncodeToString(token)
	session, ok := d.sessionStorage[h]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s session", h)
	}
	return session, nil
}

func (d *mockdb) SetSession(s *auth.Session) error {
	h := hex.EncodeToString(s.Token)
	d.sessionStorage[h] = s
	return nil
}

func (d *mockdb) RemoveSession(token []byte) error {
	h := hex.EncodeToString(token)
	if _, ok := d.sessionStorage[h]; !ok {
		return fmt.Errorf("unable to find required %s session", h)
	}
	delete(d.sessionStorage, h)
	return nil
}

func (d *mockdb) RemoveAllSessions() error {
	d.sessionStorage = make(map[string]*auth.Session)
	return nil
}
