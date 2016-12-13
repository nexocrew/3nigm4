//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package auth

import (
	"time"
)

// Level type describe available user's permission
// levels.
type Level uint

// Common levels used to identify tipical figures
// that can access a service, this list can be expanded.
const (
	LevelUser  Level = iota // common user, will not be able to administer a service;
	LevelAdmin Level = iota // administrator will be able to perform maintainance tasks.
)

// Permissions struct describe user's permisisons
// on a service basis, if the user is a sper-admin
// a special bool flag will be setted.
type Permissions struct {
	SuperAdmin bool             `bson:"superadmin,omitempty" json:"superadmin,omitempty"` // special user that have all permissions on all services;
	Services   map[string]Level `bson:"services" json:"services"`                         // permissions organised per service, the "all" can be used for generalised behaviour.
}

// User struct identify a registered
// user to the service.
type User struct {
	Username       string      `bson:"username"`           // user name;
	FullName       string      `bson:"fullname,omitempty"` // complete full name;
	HashedPassword []byte      `bson:"pwdhash"`            // hashed password;
	Email          string      `bson:"email,omitempty"`    // user's verified email;
	Permissions    Permissions `bson:"permissions"`        // the permissions associated to the user;
	IsDisabled     bool        `bson:"disabled"`           // user active (true) or not (false).
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
