// 3nigm4 s3backend package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016

// Package s3backend expose S3 interaction capabilities backended
// by the FileManager package. All operations are managed by a
// concurrent working queue and tend to be asyncronous.
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

// Session struct is composed of several private
// fields and expose the async mechanism chans.
type Session struct {
	// private vars
	workingQueue *wq.WorkingQueue
	s3           *s3.S3
	// exposed vars
	ErrorChan      chan error    // returned errors;
	UploadedChan   chan OpResult // status of uploads;
	DownloadedChan chan OpResult // return chan for async downloaded chunks;
	DeletedChan    chan OpResult // manage deletion results from wq.

}

// OpResult this struct represent the status of an async
// operation, of any type (upload, download, delete, ...).
// Not all field will be present: Error and Data properties
// will be only present if an error occurred or a download
// transaction has been required. Notice that two id are
// managed: a file id (used to identify the target file on S3)
// and RequestID used to associate a request with the async
// result produced.
type OpResult struct {
	ID        string // file id string;
	RequestID string // request (tx) id string (not file id);
	Data      []byte // downloaded data, if any;
	Error     error  // setted if an error was produced fro the upload instruction.
}

// NewSession initialise a new S3 session for the file
// storage capabilities-
func NewSession(endpoint, region, id, secret, token string, workersize, queuesize int, verbose bool) (*Session, error) {
	// get credentials
	creds := credentials.NewStaticCredentials(id, secret, token)

	// set log level
	logLevel := aws.LogOff
	if verbose == true {
		logLevel = aws.LogDebug
	}

	session := &Session{
		s3: s3.New(session.New(), &aws.Config{
			Endpoint:    &endpoint,
			Region:      &region,
			Credentials: creds,
			LogLevel:    &logLevel,
		}),
		ErrorChan:      make(chan error, workersize),
		UploadedChan:   make(chan OpResult, workersize),
		DownloadedChan: make(chan OpResult, workersize),
		DeletedChan:    make(chan OpResult, workersize),
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
// invoking the NewSession function.
func (bs *Session) Close() {
	bs.workingQueue.Close()
}

type args struct {
	// file data
	bucketName string
	id         string
	// s3
	backendSession *Session
	// put specifics
	fileType string
	fileData []byte
	expires  *time.Time
	// transaction id
	requestID string
}

func upload(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
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
		arguments.backendSession.UploadedChan <- OpResult{
			ID:        arguments.id,
			Error:     err,
			RequestID: arguments.requestID,
		}
		return err
	}

	// send result back in chan
	arguments.backendSession.UploadedChan <- OpResult{
		ID:        arguments.id,
		Error:     nil,
		RequestID: arguments.requestID,
	}
	return nil
}

func delete(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}

	// delete params
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(arguments.bucketName),
		Key:    aws.String(arguments.id),
	}

	_, err := arguments.backendSession.s3.DeleteObject(params)
	if err != nil {
		arguments.backendSession.DeletedChan <- OpResult{
			ID:        arguments.id,
			Error:     err,
			RequestID: arguments.requestID,
		}
		return err
	}

	// send deletion result back to chan
	arguments.backendSession.DeletedChan <- OpResult{
		ID:        arguments.id,
		Error:     nil,
		RequestID: arguments.requestID,
	}
	return nil
}

func download(a interface{}) error {
	var arguments *args
	var ok bool
	if arguments, ok = a.(*args); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
		return fmt.Errorf("unexpected argument type, having %s expecting *args", reflect.TypeOf(a))
	}

	// download params
	params := &s3.GetObjectInput{
		Bucket: aws.String(arguments.bucketName),
		Key:    aws.String(arguments.id),
	}

	response, err := arguments.backendSession.s3.GetObject(params)
	if err != nil {
		arguments.backendSession.DownloadedChan <- OpResult{
			ID:        arguments.id,
			Error:     err,
			RequestID: arguments.requestID,
			Data:      nil,
		}
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	// send download result back to chan
	arguments.backendSession.DownloadedChan <- OpResult{
		ID:        arguments.id,
		Error:     nil,
		RequestID: arguments.requestID,
		Data:      buf.Bytes(),
	}
	return nil
}

// Delete removes a identified file from a S3 storage bucket.
func (bs *Session) Delete(bucketName, id, requestid string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
		requestID:      requestid,
	}
	bs.workingQueue.SendJob(delete, a)
}

// Upload send a single data blob to a S3 storage.
func (bs *Session) Upload(bucketName, id, requestid string, data []byte, expires *time.Time) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		fileType:       http.DetectContentType(data),
		fileData:       data,
		expires:        expires,
		backendSession: bs,
		requestID:      requestid,
	}

	bs.workingQueue.SendJob(upload, a)
}

// Download get a single file from a S3 bucket.
func (bs *Session) Download(bucketName, id, requestid string) {
	a := &args{
		bucketName:     bucketName,
		id:             id,
		backendSession: bs,
		requestID:      requestid,
	}

	bs.workingQueue.SendJob(download, a)
}

// SaveChunks start the async upload of all argument passed chunks
// generating a single name for each one (that must be keeped in
// order to get back the file later on).
func (bs *Session) SaveChunks(filename, bucket string, chunks [][]byte, hashedValue []byte, expirets *time.Time) ([]string, error) {
	paths := make([]string, len(chunks))
	for idx, chunk := range chunks {
		id, err := fm.ChunkFileId(filename, idx, hashedValue)
		if err != nil {
			return nil, err
		}
		bs.Upload(bucket, id, id, chunk, expirets)
		paths[idx] = id
	}
	return paths, nil
}

// RetrieveChunks starts the async retrieve of previously uploaded
// chunks starting from the returned files names. The actual downloaded
// data is then returned on the DownloadedChan.
func (bs *Session) RetrieveChunks(bucket string, files []string) []string {
	for _, fname := range files {
		bs.Download(bucket, fname, fname)
	}
	return files
}
