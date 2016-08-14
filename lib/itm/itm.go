//
// 3nigm4 itm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

// Package itm (Integration Tests Manager) is intended
// to be used onlyin integration tests. Should never
// be usedin production environments or live processing.
// It should never be included by anything outside
// the "_test.go" files.
package itm

// Go default packages
import (
	"fmt"
	"os"
	"strconv"
)

var ds *IntegrationTestsDataSource // Static variable used to implement singleton pattern.

// IntegrationTestsDataSource in memory structure used to
// memorise all available testing context vars.
type IntegrationTestsDataSource struct {
	dbAddress  string // the db url or ip address;
	dbPort     int    // the db port;
	dbUserName string // the db username;
	dbPassword string // the db password;
	dbAuth     string // the db authorisation db;
	// S3 credentials and config
	s3Id       string // s3 id;
	s3Secret   string // s3 API secret key;
	s3Token    string // s3 token;
	s3Endpoint string // s3 endpoint;
	s3Region   string // s3 region;
	s3Bucket   string // s3 bucket name.
}

// Constant db values (usable in local contexts).
const (
	kDbAddress        = "127.0.0.1"
	kDbPort           = 27017
	kDbUserName       = ""
	kDbPassword       = ""
	kDbAuth           = "admin"
	kDbFullAddressFmt = "mongodb://%s:%s@%s:%d/?authSource=%s"
	kS3Id             = ""
	kS3Secret         = ""
	kS3Token          = ""
	kS3Endpoint       = "s3.amazonaws.com"
	kS3Region         = "eu-central-1"
	kS3BucketName     = "3nigm4"
)

// Environment defined keys: should be used
// in automated environment (CI tools) to
// refer to dedicated db server created on the
// fly.
const (
	kEnvDbAddr     = "ENIGM4_DB_ADDR"
	kEnvDbPort     = "ENIGM4_DB_PORT"
	kEnvDbUsr      = "ENIGM4_DB_USR"
	kEnvDbPwd      = "ENIGM4_DB_PWD"
	kEnvDbAuth     = "ENIGM4_DB_AUTH"
	kEnvS3ID       = "ENIGM4_S3_ID"
	kEnvS3Secret   = "ENIGM4_S3_SECRET"
	kEnvS3Token    = "ENIGM4_S3_TOKEN"
	kEnvS3Endpoint = "ENIGM4_S3_ENDPOINT"
	kEnvS3Region   = "ENIGM4_S3_REGION"
	kEnvS3Bucket   = "ENIGM4_S3_BUCKET"
)

// S returns an IntegrationTestsDataSource instance
// using a singleton pattern. If a shared instance
// exists returns it otherwise creates a new one,
// populates it with environment vars or costants and
// set it as shared instance.
func S() *IntegrationTestsDataSource {
	if ds == nil {
		itm := IntegrationTestsDataSource{}
		// setup vars
		env := os.Getenv(kEnvDbAddr)
		if env != "" {
			itm.dbAddress = env
		} else {
			itm.dbAddress = kDbAddress
		}
		env = os.Getenv(kEnvDbPort)
		if env != "" {
			itm.dbPort, _ = strconv.Atoi(env)
		} else {
			itm.dbPort = kDbPort
		}
		env = os.Getenv(kEnvDbUsr)
		if env != "" {
			itm.dbUserName = env
		} else {
			itm.dbUserName = kDbUserName
		}
		env = os.Getenv(kEnvDbPwd)
		if env != "" {
			itm.dbPassword = env
		} else {
			itm.dbPassword = kDbPassword
		}
		env = os.Getenv(kEnvDbAuth)
		if env != "" {
			itm.dbAuth = env
		} else {
			itm.dbAuth = kDbAuth
		}
		// S3 context
		env = os.Getenv(kEnvS3ID)
		if env != "" {
			itm.s3Id = env
		} else {
			itm.s3Id = kS3Id
		}
		env = os.Getenv(kEnvS3Secret)
		if env != "" {
			itm.s3Secret = env
		} else {
			itm.s3Secret = kS3Secret
		}
		env = os.Getenv(kEnvS3Token)
		if env != "" {
			itm.s3Token = env
		} else {
			itm.s3Token = kS3Token
		}
		env = os.Getenv(kEnvS3Endpoint)
		if env != "" {
			itm.s3Endpoint = env
		} else {
			itm.s3Endpoint = kS3Endpoint
		}
		env = os.Getenv(kEnvS3Region)
		if env != "" {
			itm.s3Region = env
		} else {
			itm.s3Region = kS3Region
		}
		env = os.Getenv(kEnvS3Bucket)
		if env != "" {
			itm.s3Bucket = env
		} else {
			itm.s3Bucket = kS3BucketName
		}
		// assign singleton
		ds = &itm
	}
	return ds
}

// DbAddress return singleton integration tests db address
// as a string.
func (i *IntegrationTestsDataSource) DbAddress() string {
	return i.dbAddress
}

// DbPort returns singleton integration tests db port
// as an integer.
func (i *IntegrationTestsDataSource) DbPort() int {
	return i.dbPort
}

// DbPortString returns singleton integration tests db port
// as a string.
func (i *IntegrationTestsDataSource) DbPortString() string {
	return strconv.Itoa(i.dbPort)
}

// DbUserName returns singleton integration tests db username
// as a string.
func (i *IntegrationTestsDataSource) DbUserName() string {
	return i.dbUserName
}

// DbPassword returns singleton integration tests db password
// as a string.
func (i *IntegrationTestsDataSource) DbPassword() string {
	return i.dbPassword
}

// DbAuth returns singleton integration tests authorisation
// db name as a string.
func (i *IntegrationTestsDataSource) DbAuth() string {
	return i.dbAuth
}

// DbFullAddress returns db full address composed with all previously
// returned elements.
func (i *IntegrationTestsDataSource) DbFullAddress() string {
	return fmt.Sprintf(kDbFullAddressFmt, i.dbUserName, i.dbPassword, i.dbAddress, i.dbPort, i.dbAuth)
}

// S3Id returns s3 id string for credentials.
func (i *IntegrationTestsDataSource) S3Id() string {
	return i.s3Id
}

// S3Secret returns s3 secret string.
func (i *IntegrationTestsDataSource) S3Secret() string {
	return i.s3Secret
}

// S3Token the s3 token value.
func (i *IntegrationTestsDataSource) S3Token() string {
	return i.s3Token
}

// S3Endpoint S3 endpoint address.
func (i *IntegrationTestsDataSource) S3Endpoint() string {
	return i.s3Endpoint
}

// S3Region the S3 region address.
func (i *IntegrationTestsDataSource) S3Region() string {
	return i.s3Region
}

// S3Bucket the S3 bucket name.
func (i *IntegrationTestsDataSource) S3Bucket() string {
	return i.s3Bucket
}
