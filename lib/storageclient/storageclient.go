// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

// Package storageclient expose client side API usage
// for the secure storage service (S3 frontend) this package
// implements the filemanager DataSaver interface.
package storageclient

// Std golang dependencies.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

const (
	jobPath     = "/v1/storage/job"
	verifySleep = 500 * time.Millisecond
)

// StorageClient is the base structure used to implement the
// interface methods.
type StorageClient struct {
	// service coordinates
	address string
	port    int
	token   string
	// working queue
	workingQueue *wq.WorkingQueue
	ErrorChan    chan error
	downloadChan chan ct.OpResult
	uplaodChan   chan ct.OpResult
	deletedChan  chan ct.OpResult
	// requests status
	requests map[string]*RequestStatus
}

// NewStorageClient creates a new StorageClient structure and
// setup all required properties. It'll start the working queue
// that'll be used to enqueue http API facing jobs.
func NewStorageClient(
	address string,
	port int,
	token string,
	workersize, queuesize int) (*StorageClient, error) {
	// creates base object
	sc := &StorageClient{
		address:      address,
		port:         port,
		token:        token,
		ErrorChan:    make(chan error, workersize),
		downloadChan: make(chan ct.OpResult, workersize),
		uplaodChan:   make(chan ct.OpResult, workersize),
		deletedChan:  make(chan ct.OpResult, workersize),
		requests:     make(map[string]*RequestStatus),
	}
	// create working queue
	sc.workingQueue = wq.NewWorkingQueue(workersize, queuesize, sc.ErrorChan)
	// startup chan management routine
	go sc.manageChans()
	// start working queue
	if err := sc.workingQueue.Run(); err != nil {
		return nil, err
	}
	return sc, nil
}

// Close close the active working queue.
func (s *StorageClient) Close() {
	s.workingQueue.Close()
}

// jobArgs standard job arguments passed to concurrent jobs while
// adding them to the working queue instance.
type jobArgs struct {
	client    *StorageClient
	args      *ct.CommandArguments
	requestID string
}

// postGenericJob implement a generic POST job operation, can be used for
// any available command.
func postGenericJob(arguments *jobArgs, command string) (*ct.JobPostResponse, error) {
	// define request body
	job := ct.JobPostRequest{
		Command:   command,
		Arguments: arguments.args,
	}
	body, err := json.Marshal(&job)
	if err != nil {
		return nil, err
	}

	// create http request
	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s:%d%s", arguments.client.address, arguments.client.port, jobPath),
		bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(ct.SecurityTokenKey, arguments.client.token)
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf(
			"service returned wrong status code: having %d expecting %d, unable to proceed",
			resp.StatusCode,
			http.StatusAccepted)
	}
	// get job ID
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var jobResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobResponse)
	if err != nil {
		return nil, err
	}

	return &jobResponse, nil
}

// getGenericJob can be used to verify any API job created with the POST
// request.
func getGenericJob(arguments *jobArgs, jobID string) (*ct.JobGetRequest, error) {
	for {
		client := &http.Client{}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s:%d%s/%s",
				arguments.client.address,
				arguments.client.port,
				jobPath,
				jobID),
			nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set(ct.SecurityTokenKey, arguments.client.token)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusAccepted:
			continue
		case http.StatusOK:
			getBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			var download ct.JobGetRequest
			err = json.Unmarshal(getBody, &download)
			if err != nil {
				return nil, err
			}
			return &download, nil
		default:
			return nil, fmt.Errorf(
				"unexpected status having %d expecting %d or %d",
				resp.StatusCode,
				http.StatusAccepted,
				http.StatusOK)
		}
		// sleep to avoid spinning on the CPU
		time.Sleep(verifySleep)
	}
}

// upload the job that'll be enqueued in the working queue to perform
// an upload.
func upload(a interface{}) error {
	var arguments *jobArgs
	var ok bool
	if arguments, ok = a.(*jobArgs); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
		return fmt.Errorf("unexpected argument type, having %s expecting *jobArgs", reflect.TypeOf(a))
	}

	// perform generic post
	postResponse, err := postGenericJob(arguments, "UPLOAD")
	if err != nil {
		arguments.client.uplaodChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}

	// loop to verify succesfull request
	getResponse, err := getGenericJob(arguments, postResponse.JobID)
	if err != nil {
		arguments.client.uplaodChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}
	if getResponse.Error != "" {
		arguments.client.uplaodChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     fmt.Errorf("%s", getResponse.Error),
		}
		return err
	}

	arguments.client.uplaodChan <- ct.OpResult{
		RequestID: arguments.requestID,
		ID:        arguments.args.ResourceID,
	}
	return nil
}

// download a file from the API frontend that'll be enqueued in the
// working queue to perform a download.
func download(a interface{}) error {
	var arguments *jobArgs
	var ok bool
	if arguments, ok = a.(*jobArgs); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
		return fmt.Errorf("unexpected argument type, having %s expecting *jobArgs", reflect.TypeOf(a))
	}

	// perform generic post
	postResponse, err := postGenericJob(arguments, "DOWNLOAD")
	if err != nil {
		arguments.client.downloadChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}

	// loop to verify succesfull request
	getResponse, err := getGenericJob(arguments, postResponse.JobID)
	if err != nil {
		arguments.client.downloadChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}
	if getResponse.Error != "" {
		arguments.client.downloadChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     fmt.Errorf("%s", getResponse.Error),
		}
		return err
	}

	arguments.client.downloadChan <- ct.OpResult{
		RequestID: arguments.requestID,
		ID:        arguments.args.ResourceID,
		Data:      getResponse.Data,
	}
	return nil
}

// remove the job that'll be enqueued in the working queue to perform
// a file deletion.
func remove(a interface{}) error {
	var arguments *jobArgs
	var ok bool
	if arguments, ok = a.(*jobArgs); !ok {
		// in this case no id can be retrieved, that's
		// why no upload response is retuned.
		return fmt.Errorf("unexpected argument type, having %s expecting *jobArgs", reflect.TypeOf(a))
	}

	// perform generic post
	postResponse, err := postGenericJob(arguments, "DELETE")
	if err != nil {
		arguments.client.deletedChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}

	// loop to verify succesfull request
	getResponse, err := getGenericJob(arguments, postResponse.JobID)
	if err != nil {
		arguments.client.deletedChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     err,
		}
		return err
	}
	if getResponse.Error != "" {
		arguments.client.deletedChan <- ct.OpResult{
			RequestID: arguments.requestID,
			ID:        arguments.args.ResourceID,
			Error:     fmt.Errorf("%s", getResponse.Error),
		}
		return err
	}

	arguments.client.deletedChan <- ct.OpResult{
		RequestID: arguments.requestID,
		ID:        arguments.args.ResourceID,
	}
	return nil
}

// SaveChunks start the async upload of all argument passed chunks
// generating a single name for each one.
func (s *StorageClient) SaveChunks(filename string, chunks [][]byte, hashedValue []byte, expire time.Duration, permission *fm.Permission) ([]string, error) {
	now := time.Now()
	requestID := generateTranscationID(filename, &now)
	// check for pending uploads
	_, ok := s.requests[requestID]
	if ok {
		return nil, fmt.Errorf("unable to proceed another job is going on with request ID %s", requestID)
	}
	s.requests[requestID] = NewRequestStatus(requestID, len(chunks))

	paths := make([]string, len(chunks))
	for idx, chunk := range chunks {
		id, err := fm.ChunkFileId(filename, idx, hashedValue)
		if err != nil {
			return nil, err
		}

		// create args struct
		commandArgs := &ct.CommandArguments{
			ResourceID: id,
			Data:       chunk,
			TimeToLive: expire,
		}
		if permission != nil {
			commandArgs.Permission = permission.Permission
			commandArgs.SharingUsers = permission.SharingUsers
		}
		ja := &jobArgs{
			client:    s,
			args:      commandArgs,
			requestID: requestID,
		}
		// add nil record to request status
		err = s.requests[requestID].SetStatus(id, false, nil)
		if err != nil {
			return nil, err
		}

		// enqueue on working queue
		s.workingQueue.SendJob(upload, ja)
		paths[idx] = id
	}

	// wait for upload to complete
	for {
		if s.requests[requestID].Completed() {
			break
		}
		time.Sleep(verifySleep)
	}

	return paths, nil
}

// RetrieveChunks starts the async retrieve of previously uploaded
// chunks starting from the returned files names.
func (s *StorageClient) RetrieveChunks(filename string, files []string) ([][]byte, error) {
	now := time.Now()
	requestID := generateTranscationID(filename, &now)
	// check for pending uploads
	_, ok := s.requests[requestID]
	if ok {
		return nil, fmt.Errorf("unable to proceed another job is going on with request ID %s", requestID)
	}
	s.requests[requestID] = NewRequestStatus(requestID, len(files))

	for _, id := range files {
		ja := &jobArgs{
			client: s,
			args: &ct.CommandArguments{
				ResourceID: id,
			},
			requestID: requestID,
		}
		// add nil record to request status
		err := s.requests[requestID].SetStatus(id, false, nil)
		if err != nil {
			return nil, err
		}

		// enqueue on working queue
		s.workingQueue.SendJob(download, ja)
	}

	// wait for download to complete
	for {
		if s.requests[requestID].Completed() {
			break
		}
		time.Sleep(verifySleep)
	}

	// geta downloaded chunks
	chunks := make([][]byte, len(files))
	for idx, id := range files {
		status, ok := s.requests[requestID].GetStatus(id)
		if !ok {
			return nil, fmt.Errorf("unable to access downloaded data chunks for resource %s", id)
		}
		if status == nil {
			return nil, fmt.Errorf("required download status info are not avalable for resource %s", id)
		}
		if status.Data == nil {
			return nil, fmt.Errorf("unable to access downloaded intenal struct for resource %s", id)
		}
		chunks[idx] = status.Data
	}
	return chunks, nil
}

// composedError compose an error from a slice of related errors.
func composedError(errors map[string]error) error {
	var errDescription string
	for key, value := range errors {
		errDescription += fmt.Sprintf("resource: %s error: %s\n", key, value)
	}
	return fmt.Errorf("founded following errors: %s", errDescription)
}

// DeleteChunks delete, requiring the API frontend, all resources
// composing a file (several resources compose a single file).
func (s *StorageClient) DeleteChunks(filename string, files []string) error {
	now := time.Now()
	requestID := generateTranscationID(filename, &now)
	// check for pending uploads
	_, ok := s.requests[requestID]
	if ok {
		return fmt.Errorf("unable to proceed another job is going on with request ID %s", requestID)
	}
	s.requests[requestID] = NewRequestStatus(requestID, len(files))

	for _, id := range files {
		ja := &jobArgs{
			client: s,
			args: &ct.CommandArguments{
				ResourceID: id,
			},
			requestID: requestID,
		}
		// add nil record to request status
		err := s.requests[requestID].SetStatus(id, false, nil)
		if err != nil {
			return err
		}

		// enqueue on working queue
		s.workingQueue.SendJob(remove, ja)
	}

	// wait for download to complete
	for {
		if s.requests[requestID].Completed() {
			break
		}
		time.Sleep(verifySleep)
	}

	// check for errors
	errors := make(map[string]error)
	for _, id := range files {
		status, ok := s.requests[requestID].GetStatus(id)
		if !ok {
			return fmt.Errorf("unable to access delete operation for resource %s", id)
		}
		if status == nil {
			return fmt.Errorf("required delete status info are not avalable for resource %s", id)
		}
		if status.Err != nil {
			errors[id] = status.Err
		}
	}
	// if any error found return a composed error
	if len(errors) != 0 {
		return composedError(errors)
	}

	return nil
}
