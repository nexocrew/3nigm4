//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package backend

// Third party libs
import (
	// https://docs.aws.amazon.com/sdk-for-go/latest/v1/developerguide/common-examples.title.html#amazon-s3
	//
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BackendSession struct {
	Bucket string
	Key    string
}

func (bs *BackendSession) Upload(data []byte) (string, error) {
	uploader := s3manager.NewUploader(session.New(&aws.Config{
		Region: aws.String("eu-west-1"),
	}))
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   data,
		Bucket: aws.String(bs.Bucket),
		Key:    aws.String(bs.Key),
	})
	if err != nil {
		return "", err
	}

	return result.UploadID, nil
}
