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
	"github.com/couchbase/gocb"
)

func startupCouchbaseConnection(cluster, bucket, password string) (*gocb.Bucket, error) {
	// connect to db
	cl, err := gocb.Connect(cluster)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to db cluster: %s.\n", err.Error())
	}
	buck, err := cl.OpenBucket(bucket, password)
	if err != nil {
		return nil, fmt.Errorf("Unable to open bucket: %s.\n", err.Error())
	}
	return buck, nil
}

/*
// Test struct
	tests := &User{
		Username:     "user",
		FullName:     "User name",
		PasswordHash: []byte("a73d92ie023"),
		Email:        "user@mail.com",
		IsDisabled:   false,
	}
	cas, err := bucket.Insert("userA", &tests, 0)
	if err != nil {
		return fmt.Errorf("unable to insert document: %s", err.Error())

	}
	log.MessageLog("Doc inserted! %v.\n", cas)
*/
