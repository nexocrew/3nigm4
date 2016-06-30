//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package s3backend

import (
	"bytes"
	"fmt"
	"reflect"
	"time"
)

// Third party libs
import (
	// https://docs.aws.amazon.com/sdk-for-go/latest/v1/developerguide/common-examples.title.html#amazon-s3
	//
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/awserr"
	_ "github.com/aws/aws-sdk-go/aws/awsutil"
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
	// exposed vars
	ErrorChan  chan error
	SendedChan chan *s3.PutObjectOutput
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
			Endpoint:    &endpoint,
			Region:      &region,
			Credentials: creds,
			LogLevel:    &logLevel,
		},
		ErrorChan:  make(chan error, workersize),
		SendedChan: make(chan *s3.PutObjectOutput, workersize),
	}

	// create working queue
	session.workingQueue = wq.NewWorkingQueue(workersize, queuesize, session.ErrorChan)
	if err := session.workingQueue.Run(); err != nil {
		return nil, err
	}
	return session, nil
}

func (bs *S3BackendSession) Close() {
	bs.workingQueue.Close()
}

type args struct {
	// file data
	bucketName string
	id         string
	fileType   string
	fileData   []byte
	expires    *time.Time
	// s3 config
	config         *aws.Config
	backendSession *S3BackendSession
}

func upload(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}
	svc := s3.New(session.New(), arguments.config)

	// compose params
	params := &s3.PutObjectInput{
		Bucket:        aws.String(arguments.bucketName), // required
		Key:           aws.String(arguments.id),         // required
		ACL:           aws.String("private"),
		Body:          bytes.NewReader(arguments.fileData),
		ContentLength: aws.Int64(int64(len(arguments.fileData))),
		ContentType:   aws.String(arguments.fileType),
		Expires:       arguments.expires,
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), //required
		},
	}
	result, err := svc.PutObject(params)
	if err != nil {
		return err
	}

	// send result back in chan
	arguments.backendSession.SendedChan <- result

	return nil
}

func (bs *S3BackendSession) Upload(bucketName, id, fileType string, data []byte, expires *time.Time) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		fileType:       fileType,
		fileData:       data,
		expires:        expires,
		config:         bs.config,
		backendSession: bs,
	}

	bs.workingQueue.SendJob(upload, a)
}
