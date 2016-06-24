//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package backend

import (
	"bytes"
)

// Third party libs
import (
	// https://docs.aws.amazon.com/sdk-for-go/latest/v1/developerguide/common-examples.title.html#amazon-s3
	//
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Internal dependencies
import (
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

type S3BackendSession struct {
	// private vars
	config       *aws.Config
	workingQueue *wq.WorkingQueue
	errorChan    chan error
}

func NewS3BackendSession(endpoint, region, id, secret, token string, workersize, queuesize int, verbose bool) (*S3BackendSession, error) {
	// get credentials
	creds := credentials.NewStaticCredentials(id, secret, token)

	// set log level
	logLevel := aws.LogOff
	if verbose == true {
		logLevel = aws.LogDebug
	}

	session := &S3BackendSession{
		config: &aws.Config{
			Endpoint:         endpoint,
			Region:           region,
			S3ForcePathStyle: true,
			Credentials:      creds,
			LogLevel:         logLevel,
		},
		errorChan: make(chan error),
	}
	// create working queue
	session.workingQueue = wq.NewWorkingQueue(workersize, queuesize, session.errorChan)
	if err := session.workingQueue.Run(); err != nil {
		return nil, err
	}
	return session, nil
}

func (bs *S3BackendSession) Upload(data []byte) (string, error) {
	uploader := s3manager.NewUploader(session.New(&aws.Config{
		Region: aws.String("eu-west-1"),
	}))

	buf := bytes.NewBuffer(data)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   buf,
		Bucket: aws.String(bs.Bucket),
		Key:    aws.String(bs.Key),
	})
	if err != nil {
		return "", err
	}

	return result.UploadID, nil
}
