//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

import (
	"time"
)

// User struct identify a registered
// user to the service.
type User struct {
	Username       string `bson:"username" `          // user name;
	FullName       string `bson:"fullname,omitempty"` // complete full name;
	HashedPassword []byte `bson:"pwdhash"`            // hashed password;
	Email          string `bson:"email,omitempty"`    // user's verified email;
	IsDisabled     bool   `bson:"disabled"`           // user active (true) or not (false).
}

// Session contains information about loggedin
// for authenticated users.
type Session struct {
	Token        []byte        `bson:"token"`       // token for the session;
	Username     string        `bson:"username"`    // username associated to session;
	LoginTime    time.Time     `bson:"login_ts"`    // timestamp of login time for this session;
	LastSeenTime time.Time     `bson:"lastseen_ts"` // last call to an API done by the user;
	TimeToLive   time.Duration `bson:"timetolive"`  // time of validity of the session.
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
	dbclient database
	// service
	address string
	port    int
}
