//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"os"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/logger"
)

// Third party libs
import (
	"github.com/couchbase/gocb"
	_ "github.com/spf13/cobra"
)

var log *logger.LogFacility

type User struct {
	Username     string
	FullName     string
	PasswordHash []byte
	Email        string
	IsDisabled   bool
}

func main() {

	// start up logging facility
	log = logger.NewLogFacility("authserver", true, true)

	cluster, err := gocb.Connect("couchbase://localhost")
	if err != nil {
		log.CriticalLog("Unable to connect to db cluster: %s.\n", err.Error())
		os.Exit(1)
	}
	bucket, err := cluster.OpenBucket("3nigm4", "")
	if err != nil {
		log.CriticalLog("Unable to open bucket: %s.\n", err.Error())
		os.Exit(1)
	}
	log.MessageLog("Bucket opened: %v.\n", bucket)

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
		log.CriticalLog("Unable to insert document: %s.\n", err.Error())
		os.Exit(1)
	}
	log.MessageLog("Doc inserted! %v.\n", cas)

	os.Exit(0)
}
