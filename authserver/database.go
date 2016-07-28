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

func mgoSession(address, bucket, user, password, auth string) (*mgo.Session, error) {
	// connect to db
	return mgo.Dial(fmt.Sprintf(format, ...))
}
