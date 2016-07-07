//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package s3backend

import (
	"bytes"
	"fmt"
	"net/http"
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
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

type S3BackendSession struct {
	// private vars
	workingQueue *wq.WorkingQueue
	s3           *s3.S3
	// exposed vars
	ErrorChan      chan error
	UploadedChan   chan string
	DownloadedChan chan DownloadRequest
}

type DownloadRequest struct {
	Data      []byte
	RequestId string
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
		s3: s3.New(session.New(), &aws.Config{
			Endpoint:    &endpoint,
			Region:      &region,
			Credentials: creds,
			LogLevel:    &logLevel,
		}),
		ErrorChan:      make(chan error, workersize),
		UploadedChan:   make(chan string, workersize),
		DownloadedChan: make(chan DownloadRequest, workersize),
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
	// s3
	backendSession *S3BackendSession
	// put specifics
	fileType string
	fileData []byte
	expires  *time.Time
}

func upload(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}

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
	_, err := arguments.backendSession.s3.PutObject(params)
	if err != nil {
		return err
	}

	// send result back in chan
	arguments.backendSession.UploadedChan <- arguments.id
	return nil
}

func delete(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}

	// delete params
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(arguments.bucketName),
		Key:    aws.String(arguments.id),
	}

	_, err := arguments.backendSession.s3.DeleteObject(params)
	if err != nil {
		return err
	}

	return nil
}

func download(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}

	// download params
	params := &s3.GetObjectInput{
		Bucket: aws.String(arguments.bucketName),
		Key:    aws.String(arguments.id),
	}

	response, err := arguments.backendSession.s3.GetObject(params)
	if err != nil {
		return nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	arguments.backendSession.DownloadedChan <- DownloadRequest{
		Data:      buf.Bytes(),
		RequestId: arguments.id,
	}

	return nil

}

func (bs *S3BackendSession) Delete(bucketName, id string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
	}
	bs.workingQueue.SendJob(delete, a)
}

func (bs *S3BackendSession) Upload(bucketName, id string, data []byte, expires *time.Time) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		fileType:       http.DetectContentType(data),
		fileData:       data,
		expires:        expires,
		backendSession: bs,
	}

	bs.workingQueue.SendJob(upload, a)
}

func (bs *S3BackendSession) Download(bucketName, id string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
	}

	bs.workingQueue.SendJob(download, a)
}

func (bs *S3BackendSession) SaveChunks(filename string, chunks [][]byte, hashedValue []byte) ([]string, error) {
	paths := make([]string, len(chunks))
	for idx, chunk := range chunks {
		id, err := fm.ChunkFileId(filename, idx, hashedValue)
		if err != nil {
			return nil, err
		}
		bs.Upload(bucketName, id, data, expires)
		paths[idx] = id
	}
	return paths, nil
}

func (bs *S3BackendSession) RetrieveChunks(files []string) ([][]byte, error) {

}
