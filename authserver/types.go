//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

import (
	"time"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
)

// User struct identify a registered
// user to the service.
type User struct {
	Username       string `json:"username"` // user name;
	FullName       string `json:"fullname"` // complete full name;
	HashedPassword []byte `json:"pwdhash"`  // hashed password;
	Email          string `json:"email"`    // user's verified email;
	IsDisabled     bool   `json:"disabled"` // user active (true) or not (false).
}

// Session contains information about loggedin
// for authenticated users.
type Session struct {
	Token        []byte    `json:"token"`       // token for the session;
	Username     string    `json:"username"`    // username associated to session;
	LoginTime    time.Time `json:"login_ts"`    // timestamp of login time for this session;
	LastSeenTime time.Time `json:"lastseen_ts"` // last call to an API done by the user.
}

// Arguments management struct.
type args struct {
	// server basic args
	verbose bool
	colored bool
	// mongodb
	dbAddresses string
	dbUsername  string
	dbPassword  string
	dbAuth      string
	// runtime allocated
	session *mgo.Session
	// service
	address string
	port    int
}
