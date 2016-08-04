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
package auth

// Golang std libs
import (
	"encoding/hex"
	"fmt"
	_ "testing"
)

type mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	userStorage    map[string]*User
	sessionStorage map[string]*Session
}

func newMockDb(args *DbArgs) *mockdb {
	return &mockdb{
		addresses:      composeDbAddress(args),
		user:           args.User,
		password:       args.Password,
		authDb:         args.AuthDb,
		userStorage:    make(map[string]*User),
		sessionStorage: make(map[string]*Session),
	}
}

func (d *mockdb) Copy() Database {
	return d
}

func (d *mockdb) Close() {
}

func (d *mockdb) GetUser(username string) (*User, error) {
	user, ok := d.userStorage[username]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s user", username)
	}
	return user, nil
}

func (d *mockdb) SetUser(user *User) error {
	_, ok := d.userStorage[user.Username]
	if ok {
		return fmt.Errorf("user %s already exist in the db", user.Username)
	}
	d.userStorage[user.Username] = user
	return nil
}

func (d *mockdb) RemoveUser(username string) error {
	delete(d.userStorage, username)
	return nil
}

func (d *mockdb) GetSession(token []byte) (*Session, error) {
	h := hex.EncodeToString(token)
	session, ok := d.sessionStorage[h]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s session", h)
	}
	return session, nil
}

func (d *mockdb) SetSession(s *Session) error {
	h := hex.EncodeToString(s.Token)
	d.sessionStorage[h] = s
	return nil
}

func (d *mockdb) RemoveSession(token []byte) error {
	h := hex.EncodeToString(token)
	delete(d.sessionStorage, h)
	return nil
}

func (d *mockdb) RemoveAllSessions() error {
	d.sessionStorage = make(map[string]*Session)
	return nil
}
