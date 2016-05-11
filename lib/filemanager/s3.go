//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import ()

// Third party packages
import (
	"github.com/mitchellh/goamz/aws"
	_ "github.com/mitchellh/goamz/s3"
)

type s3FileManager struct {
	// authorisation informations
	auth aws.Auth
	// address location
	region aws.Region
}

// WriteFile upload a file to an S3 instance after dividing it
// in chunks and encrypting each of them with a ramdomly
// generated key (that is returned by the function).
func (s *s3FileManager) WriteFile(data []byte) ([]byte, error) {
	return nil, nil
}
