// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

// Package storageapiclient expose client side API usage
// for the secure storage service (S3 frontend) this package
// implements the filemanager DataSaver interface.
package storageclient

// Std golang dependencies.
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// Progress is used to maintain the status of various files
// update, it report the initial number of files and as soon
// as upload successfully terminates, the number of uploaded
// parts.
type Progress struct {
	Filename      string
	HashedValue   []byte
	NumberOfFiles int
	Progress      wg.AtomicCounter
}

// StorageClient is the base structure used to implement the
// interface methods.
type StorageClient struct {
	// service coordinates
	address string
	port    int
	token   string
	// progress montoring logic
	progresses map[string]Progress
}

// NewStorageClient creates a new StorageClient structure and
// setup all required properties.
func NewStorageClient(address string, port int, token string) *StorageClient {
	return &StorageClient{
		address: address,
		port:    port,
		token:   token,
	}
}

// Upload send a file to the API frontend to upload it to the S3
// storage. This function should be used in a background routine to
// avoid blocking the main thread.
func (s *StorageClient) Upload(args *ct.CommandArguments) (string, error) {
	// define request body
	job := ct.JobPostRequest{
		Command:   "UPLOAD",
		Arguments: args,
	}
	body, err := json.Marshal(&jobUpload)
	if err != nil {
		return "", err
	}

	// create http request
	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s:%d%s", s.address, s.port, jobPath),
		bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set(ct.SecurityTokenKey, s.token)
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("service returned wrong status code: having %d expecting %d, unable to proceed", resp.StatusCode, http.StatusAccepted)
	}
	// get job ID
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var jobResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobResponse)
	if err != nil {
		return "", err
	}

	// loop to verify succesfull request
	for {
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s:%d%s/%s",
				s.address,
				s.port,
				jobPath,
				jobResponse.JobID),
			nil)
		if err != nil {
			return "", err
		}
		req.Header.Set(ct.SecurityTokenKey, s.token)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusAccepted:
			continue
		case http.StatusOK:
			getBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			var download ct.JobGetRequest
			err = json.Unmarshal(getBody, &download)
			if download.Error != "" {
				return "", fmt.Errorf("%s", download.Error)
			}
			return args.ResourceID, nil
		default:
			return "", fmt.Errorf("unexpected status having %d expecting %d or %d", resp.StatusCode, http.StatusAccepted, http.StatusOK)
		}
		// sleep to avoid spinning on the CPU
		time.Sleep(verifySleep)
	}
}

// Download download a file from the API frontend.
// This function should be used in a background routine to
// avoid blocking the main thread.
func (s *StorageClient) Download(args *ct.CommandArguments) ([]byte, error) {
	// define request body
	job := ct.JobPostRequest{
		Command:   "DOWNLOAD",
		Arguments: args,
	}
	body, err := json.Marshal(&jobUpload)
	if err != nil {
		return nil, err
	}

	// create http request
	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s:%d%s", s.address, s.port, jobPath),
		bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(ct.SecurityTokenKey, s.token)
	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("service returned wrong status code: having %d expecting %d, unable to proceed", resp.StatusCode, http.StatusAccepted)
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

	// loop to verify succesfull request
	for {
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s:%d%s/%s",
				s.address,
				s.port,
				jobPath,
				jobResponse.JobID),
			nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set(ct.SecurityTokenKey, s.token)
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
			if download.Error != "" {
				return nil, fmt.Errorf("%s", download.Error)
			}
			return download.Data, nil
		default:
			return nil, fmt.Errorf("unexpected status having %d expecting %d or %d", resp.StatusCode, http.StatusAccepted, http.StatusOK)
		}
		// sleep to avoid spinning on the CPU
		time.Sleep(verifySleep)
	}
}

// SaveChunks start the async upload of all argument passed chunks
// generating a single name for each one.
func (s *StorageClient) SaveChunks(filename string, chunks [][]byte, hashedValue []byte, expirets *time.Time, permission *fm.Permission) ([]string, error) {
	paths := make([]string, len(chunks))
	for idx, chunk := range chunks {
		id, err := fm.ChunkFileId(filename, idx, hashedValue)
		if err != nil {
			return nil, err
		}
		args := &ct.CommandArguments{
			ResourceID: id,
			Data:       chunk,
		}
		if expirets != nil {
			args.TimeToLive = expirets.Sub(time.Now())
		}
		if permission != nil {
			args.Permission = permission.Permission
			args.SharingUsers = permission.SharedUsers
		}
		resourceID, error := s.Upload(args)
		if err != nil {
			return nil, err
		}
		paths[idx] = id
	}
	return paths, nil
}

// RetrieveChunks starts the async retrieve of previously uploaded
// chunks starting from the returned files names.
func (s *StorageClient) RetrieveChunks(files []string) ([][]byte, error) {
	for _, fname := range files {
		s.Download(&ct.CommandArguments{})
	}
	return files
}
