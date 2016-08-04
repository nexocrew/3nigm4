//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// This package is intended to be used only
// in integration tests. Should never be used
// in production environments or live processing.
// It should never be included by anything outside
// the "_test.go" files.

// itm: Integration Tests Manager.
package itm

// Go default packages
import (
	"fmt"
	"os"
	"strconv"
)

var ds *IntegrationTestsDataSource // Static variable used to implement singleton pattern.

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
	ENV_DBADDR     = "ENIGM4_DB_ADDR"
	ENV_DBPORT     = "ENIGM4_DB_PORT"
	ENV_DBUSR      = "ENIGM4_DB_USR"
	ENV_DBPWD      = "ENIGM4_DB_PWD"
	ENV_DBAUTH     = "ENIGM4_DB_AUTH"
	ENV_S3ID       = "ENIGM4_S3_ID"
	ENV_S3SECRET   = "ENIGM4_S3_SECRET"
	ENV_S3TOKEN    = "ENIGM4_S3_TOKEN"
	ENV_S3ENDPOINT = "ENIGM4_S3_ENDPOINT"
	ENV_S3REGION   = "ENIGM4_S3_REGION"
	ENV_S3BUCKET   = "ENIGM4_S3_BUCKET"
)

// Returns an IntegrationTestsDataSource instance
// using a singleton pattern. If a shared instance
// exists returns it otherwise creates a new one,
// populates it with environment vars or costants and
// set it as shared instance.
func S() *IntegrationTestsDataSource {
	if ds == nil {
		itm := IntegrationTestsDataSource{}
		// setup vars
		env := os.Getenv(ENV_DBADDR)
		if env != "" {
			itm.dbAddress = env
		} else {
			itm.dbAddress = kDbAddress
		}
		env = os.Getenv(ENV_DBPORT)
		if env != "" {
			itm.dbPort, _ = strconv.Atoi(env)
		} else {
			itm.dbPort = kDbPort
		}
		env = os.Getenv(ENV_DBUSR)
		if env != "" {
			itm.dbUserName = env
		} else {
			itm.dbUserName = kDbUserName
		}
		env = os.Getenv(ENV_DBPWD)
		if env != "" {
			itm.dbPassword = env
		} else {
			itm.dbPassword = kDbPassword
		}
		env = os.Getenv(ENV_DBAUTH)
		if env != "" {
			itm.dbAuth = env
		} else {
			itm.dbAuth = kDbAuth
		}
		// S3 context
		env = os.Getenv(ENV_S3ID)
		if env != "" {
			itm.s3Id = env
		} else {
			itm.s3Id = kS3Id
		}
		env = os.Getenv(ENV_S3SECRET)
		if env != "" {
			itm.s3Secret = env
		} else {
			itm.s3Secret = kS3Secret
		}
		env = os.Getenv(ENV_S3TOKEN)
		if env != "" {
			itm.s3Token = env
		} else {
			itm.s3Token = kS3Token
		}
		env = os.Getenv(ENV_S3ENDPOINT)
		if env != "" {
			itm.s3Endpoint = env
		} else {
			itm.s3Endpoint = kS3Endpoint
		}
		env = os.Getenv(ENV_S3REGION)
		if env != "" {
			itm.s3Region = env
		} else {
			itm.s3Region = kS3Region
		}
		env = os.Getenv(ENV_S3BUCKET)
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

// Return singleton integration tests db address
// as a string.
func (i *IntegrationTestsDataSource) DbAddress() string {
	return i.dbAddress
}

// Returns singleton integration tests db port
// as an integer.
func (i *IntegrationTestsDataSource) DbPort() int {
	return i.dbPort
}

// Returns singleton integration tests db port
// as a string.
func (i *IntegrationTestsDataSource) DbPortString() string {
	return strconv.Itoa(i.dbPort)
}

// Returns singleton integration tests db username
// as a string.
func (i *IntegrationTestsDataSource) DbUserName() string {
	return i.dbUserName
}

// Returns singleton integration tests db password
// as a string.
func (i *IntegrationTestsDataSource) DbPassword() string {
	return i.dbPassword
}

// Returns singleton integration tests authorisation
// db name as a string.
func (i *IntegrationTestsDataSource) DbAuth() string {
	return i.dbAuth
}

// Returns db full address composed with all previously
// returned elements.
func (i *IntegrationTestsDataSource) DbFullAddress() string {
	return fmt.Sprintf(kDbFullAddressFmt, i.dbUserName, i.dbPassword, i.dbAddress, i.dbPort, i.dbAuth)
}

// Returns s3 id string for credentials.
func (i *IntegrationTestsDataSource) S3Id() string {
	return i.s3Id
}

// Returns s3 secret string.
func (i *IntegrationTestsDataSource) S3Secret() string {
	return i.s3Secret
}

// The s3 token value.
func (i *IntegrationTestsDataSource) S3Token() string {
	return i.s3Token
}

// S3 endpoint address.
func (i *IntegrationTestsDataSource) S3Endpoint() string {
	return i.s3Endpoint
}

// The S3 region address.
func (i *IntegrationTestsDataSource) S3Region() string {
	return i.s3Region
}

// The S3 bucket name.
func (i *IntegrationTestsDataSource) S3Bucket() string {
	return i.s3Bucket
}
