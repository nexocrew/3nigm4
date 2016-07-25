//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// This package expose S3 interaction capabilities backended
// by the FileManager package. All operations are managed by a
// concurrent working queue and tend to be asyncronous.
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
	// Inspired by:
	// https://docs.aws.amazon.com/sdk-for-go/latest/v1/developerguide/common-examples.title.html#amazon-s3
	// See the link for more examples.
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Internal dependencies
import (
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

// The Backend session struct is composed of several private
// fields and expose the async mechanism chans.
type S3BackendSession struct {
	// private vars
	workingQueue *wq.WorkingQueue
	s3           *s3.S3
	// exposed vars
	ErrorChan      chan error           // returned errors;
	UploadedChan   chan string          // id of completed uploads;
	DownloadedChan chan DownloadRequest // return chan for async downloaded chunks.
}

// Returned by asyng download routines should be matching the
// requested id field.
type DownloadRequest struct {
	Data      []byte // actual data;
	RequestId string // unique id for the retrieved file.
}

// NewS3BackendSession initialise a new S3 session for the file
// storage capabilities-
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

// Close close the actual opened connection with S3 storage,
// should normally be used with the defer keyword after
// invoking the NewS3BackendSession function.
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

// Delete removes a identified file from a S3 storage bucket.
func (bs *S3BackendSession) Delete(bucketName, id string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
	}
	bs.workingQueue.SendJob(delete, a)
}

// Upload send a single data blob to a S3 storage.
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

// Download get a single file from a S3 bucket.
func (bs *S3BackendSession) Download(bucketName, id string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
	}

	bs.workingQueue.SendJob(download, a)
}

// SaveChunks start the async upload of all argument passed chunks
// generating a single name for each one (that must be keeped in
// order to get back the file later on).
func (bs *S3BackendSession) SaveChunks(filename, bucket string, chunks [][]byte, hashedValue []byte, expirets *time.Time) ([]string, error) {
	paths := make([]string, len(chunks))
	for idx, chunk := range chunks {
		id, err := fm.ChunkFileId(filename, idx, hashedValue)
		if err != nil {
			return nil, err
		}
		bs.Upload(bucket, id, chunk, expirets)
		paths[idx] = id
	}
	return paths, nil
}

// RetrieveChunks starts the async retrieve of previously uploaded
// chunks starting from the returned files names. The actual downloaded
// data is then returned on the DownloadedChan.
func (bs *S3BackendSession) RetrieveChunks(bucket string, files []string) []string {
	for _, fname := range files {
		bs.Download(bucket, fname)
	}
	return files
}
