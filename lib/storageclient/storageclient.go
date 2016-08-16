// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

// Package storageapiclient expose client side API usage
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

// delete the job that'll be enqueued in the working queue to perform
// a file deletion.
func delete(a interface{}) error {
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

func (s *StorageClient) updateUploadRequestStatus(uploaded ct.OpResult) {
	// upload request
	value, ok := s.requests[uploaded.RequestID]
	if !ok {
		s.ErrorChan <- fmt.Errorf("unable to find request status manager for %s", uploaded.RequestID)
		return
	}
	err := value.SetStatus(uploaded.ID, true, &uploaded)
	if err != nil {
		s.ErrorChan <- err
		return
	}
}

// manageChans manages chan messages from working queue all recived
// messages must be remapped on uplaod requests.
func (s *StorageClient) manageChans() {
	var uploadedcClosed, downloadedcClosed, deletedcClosed bool
	for {
		if uploadedcClosed == true {
			return
		}
		if downloadedcClosed == true {
			return
		}
		if deletedcClosed == true {
			return
		}
		// select on channels
		select {
		case uploaded, uploadedcOk := <-s.uplaodChan:
			if !uploadedcOk {
				uploadedcClosed = true
			} else {
				go s.updateUploadRequestStatus(uploaded)
			}
		case downloaded, downloadedcOk := <-s.downloadChan:
			if !downloadedcOk {
				downloadedcClosed = true
			} else {
				// TODO: implement it
				// go updateDownloadRequestStatus(downloaded)
			}
		case deleted, deletedcOk := <-s.deletedChan:
			if !deletedcOk {
				deletedcClosed = true
			} else {
				// TODO: implement it
				// go updateDeleteRequestStatus(deleted)
			}
		}
	}
}

// SaveChunks start the async upload of all argument passed chunks
// generating a single name for each one.
func (s *StorageClient) SaveChunks(filename string, chunks [][]byte, hashedValue []byte, expirets *time.Time, permission *fm.Permission) ([]string, error) {
	now := time.Now()
	requestID := generateTranscationId(filename, &now)
	// check for pending uploads
	_, ok := s.requests[requestID]
	if ok {
		return nil, fmt.Errorf("unable to proceed another job is uploading %s file", requestID)
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
		}
		if expirets != nil {
			commandArgs.TimeToLive = expirets.Sub(time.Now())
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
		s.requests[requestID].SetStatus(id, false, nil)

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
func (s *StorageClient) RetrieveChunks(files []string) ([][]byte, error) {
	chunks := make([][]byte, len(files))
	for idx, id := range files {
		/*
			// TODO: implement it with wq
			data, err := s.Download(&ct.CommandArguments{
				ResourceID: id,
			})
			if err != nil {
				return nil, err
			}
			chunks[idx] = data
		*/
	}
	return chunks, nil
}
