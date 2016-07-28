//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"fmt"
)

// Third party libs
import (
	"gopkg.in/mgo.v2"
)

type dbArgs struct {
	addresses []string
	user      string
	password  string
	authDb    string
}

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

func mgoSession(args *dbArgs) (*mgo.Session, error) {
	// connect to db
	return mgo.Dial(composeDbAddress(args))
}
