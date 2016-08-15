//
// 3nigm4 storageservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Internal dependencies.
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

const (
	fileContent = `Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.
		Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.`
)

var (
	mockServiceAddress = "127.0.0.1"
	mockServicePort    = 17973
	fileID             = ""
)

func TestMain(m *testing.M) {
	// start up logging facility
	log = logger.NewLogFacility("storageservice", true, true)

	// create resource name
	rawID, err := ct.RandomBytesForLen(32)
	if err != nil {
		log.CriticalLog("Unable to create random resource name: %s.\n", err.Error())
		os.Exit(1)
	}
	fileID = hex.EncodeToString(rawID)

	arguments = args{
		verbose:            true,
		colored:            true,
		dbAddresses:        fmt.Sprintf("%s:%d", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:         itm.S().DbUserName(),
		dbPassword:         itm.S().DbPassword(),
		dbAuth:             itm.S().DbAuth(),
		address:            mockServiceAddress,
		port:               mockServicePort,
		s3Endpoint:         itm.S().S3Endpoint(),
		s3Region:           itm.S().S3Region(),
		s3Id:               itm.S().S3Id(),
		s3Secret:           itm.S().S3Secret(),
		s3Token:            itm.S().S3Token(),
		s3Bucket:           itm.S().S3Bucket(),
		s3QueueSize:        200,
		s3WorkingQueueSize: 12,
	}
	databaseStartup = mockDbStartup
	authClientStartup = mockAuthStartup

	var errorCounter wq.AtomicCounter
	errorChan := make(chan error, 0)
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-errorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()
	// startup service
	go func(ec chan error) {
		err := serve(ServeCmd, nil)
		if err != nil {
			ec <- err
			return
		}
	}(errorChan)
	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (especially to verify that
	// no error is returned by concurrent server
	// startup). 3 seconds is an arbitrary, experimentally
	// defined, time on some slower systems it can be not
	// enought.
	ticker := time.Tick(3 * time.Second)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	// infinite loop:
	for {
		if timeoutCounter.Value() != 0 {
			break
		}
		if errorCounter.Value() != 0 {
			log.ErrorLog("Error returned: %s.\n", lastError)
			os.Exit(1)
		}
		time.Sleep(50 * time.Millisecond)
	}

	os.Exit(m.Run())
}

func mockDbStartup(arguments *args) (database, error) {
	mockdb := newMockDb(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})

	log.MessageLog("Mockdb %s successfully connected.\n", arguments.dbAddresses)

	return mockdb, nil
}

func mockAuthStartup(a *args) (AuthClient, error) {
	client, err := newAuthMock()
	if err != nil {
		return nil, err
	}

	log.MessageLog("Initialised mock auth service.\n")

	return client, nil
}

func TestPing(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s:%d/v1/ping", mockServiceAddress, mockServicePort),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to ping the service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}
}

func TestLoginAndLogout(t *testing.T) {
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)
	if err != nil {
		t.Fatalf("Unable to marshal request body: %s.\n", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	req, err = http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the logout request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform logout request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to perform logout, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var invalidated ct.LogoutResponse

	err = json.Unmarshal(respBody, &invalidated)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if invalidated.Invalidated == "" ||
		len(invalidated.Invalidated) == 0 {
		t.Fatalf("Invalid invalidated field: should not be nil.\n")
	}
	resp.Body.Close()

	if session.Token != invalidated.Invalidated {
		t.Fatalf("Expecting invalidating %s but found %s.\n", session.Token, invalidated.Invalidated)
	}
}

func verifyJobCompletion(t *testing.T, jobID, token string, timeout time.Duration) []byte {
	// create error chan
	var errorCounter wq.AtomicCounter
	errorChan := make(chan error, 0)
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-errorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()
	// define timeout value
	ticker := time.Tick(timeout)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	// define background call to the get API
	completeCounter := wq.AtomicCounter{}
	var getBody []byte
	go func() {
		for {
			client := &http.Client{}
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("http://%s:%d/v1/storage/job/%s", mockServiceAddress, mockServicePort, jobID),
				nil)
			if err != nil {
				errorChan <- err
			}
			req.Header.Set(ct.SecurityTokenKey, token)
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- err
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusAccepted:
				continue
			case http.StatusOK:
				completeCounter.Add(1)
				getBody, _ = ioutil.ReadAll(resp.Body)
				return
			default:
				errorChan <- fmt.Errorf("unexpected status having %d expecting %d or %d", resp.StatusCode, http.StatusAccepted, http.StatusOK)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	// infinite loop:
	for {
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Timeout reached before reciving ACK for the completion of the job.\n")
		}
		if errorCounter.Value() != 0 {
			t.Fatalf("Error encountered: %s.\n", lastError.Error())
		}
		if completeCounter.Value() != 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return getBody
}

func TestStorageResourceFlow(t *testing.T) {
	// Login the user
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)
	if err != nil {
		t.Fatalf("Unable to marshal request body: %s.\n", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	if fileID == "" {
		t.Fatalf("Resource name cannot be nil.\n")
	}

	// define request body
	jobUpload := ct.JobPostRequest{
		Command: "UPLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
			Data:       []byte(fileContent),
			Permission: Public,
		},
	}
	body, err = json.Marshal(&jobUpload)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to submit resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)
	var jobUploadResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobUploadResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobUploadResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedBody := verifyJobCompletion(t, jobUploadResponse.JobID, session.Token, 30*time.Second)

	// get response body
	var completedUpload ct.JobGetRequest
	err = json.Unmarshal(completedBody, &completedUpload)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}
	if completedUpload.Complete != true {
		t.Fatalf("Unexpected Complete flag setted to false.\n")
	}
	if completedUpload.Data != nil {
		t.Fatalf("Unexpected Data not nil in upload command.\n")
	}
	if completedUpload.Error != "" {
		t.Fatalf("Unexpected Error should be nil but found %s.\n", completedUpload.Error)
	}
	if completedUpload.CheckSum.Hash != nil {
		t.Fatalf("Unexpected checksum, should be nil.\n")
	}
	if completedUpload.CheckSum.Type != "" {
		t.Fatalf("Unexpected checksum type, should be nil but found %s.\n", completedUpload.CheckSum.Type)
	}

	// donwload resource
	jobDownload := ct.JobPostRequest{
		Command: "DOWNLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
		},
	}
	body, err = json.Marshal(&jobDownload)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}
	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to retrieve resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)

	var jobDownloadResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobDownloadResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobDownloadResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedDownloadBody := verifyJobCompletion(t, jobDownloadResponse.JobID, session.Token, 30*time.Second)
	// get response body
	var completedDownload ct.JobGetRequest
	err = json.Unmarshal(completedDownloadBody, &completedDownload)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}
	if completedDownload.Complete != true {
		t.Fatalf("Unexpected Complete flag setted to false.\n")
	}
	if completedDownload.Data == nil ||
		len(completedDownload.Data) == 0 {
		t.Fatalf("Data field should be not nil.\n")
	}
	if bytes.Compare(completedDownload.Data, []byte(fileContent)) != 0 {
		t.Fatalf("Downloaded data is not euqal to reference: %v != %v.\n", completedDownload.Data, []byte(fileContent))
	}
	if completedDownload.Error != "" {
		t.Fatalf("Unexpected Error should be nil but found %s.\n", completedDownload.Error)
	}
	if completedDownload.CheckSum.Type != "SHA256" {
		t.Fatalf("Unexpected checksum type, should be \"SHA256\" but found %s.\n", completedDownload.CheckSum.Type)
	}
	if completedDownload.CheckSum.Hash == nil {
		t.Fatalf("Unexpected nil checksum.\n")
	}

	// delete resource
	jobDelete := ct.JobPostRequest{
		Command: "DELETE",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
		},
	}
	body, err = json.Marshal(&jobDelete)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}
	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to delete resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)

	var jobDeleteResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobDeleteResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobDeleteResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedDeleteBody := verifyJobCompletion(t, jobDeleteResponse.JobID, session.Token, 30*time.Second)
	// get response body
	var completedDelete ct.JobGetRequest
	err = json.Unmarshal(completedDeleteBody, &completedDelete)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}
	if completedDelete.Complete != true {
		t.Fatalf("Unexpected Complete flag setted to false.\n")
	}
	if completedDelete.Data != nil {
		t.Fatalf("Unexpected Data not nil in upload command.\n")
	}
	if completedDelete.Error != "" {
		t.Fatalf("Unexpected Error should be nil but found %s.\n", completedDelete.Error)
	}
	if completedDelete.CheckSum.Hash != nil {
		t.Fatalf("Unexpected checksum, should be nil.\n")
	}
	if completedDelete.CheckSum.Type != "" {
		t.Fatalf("Unexpected checksum type, should be nil but found %s.\n", completedDelete.CheckSum.Type)
	}
}

func TestStorageGetResourceNotAuthenticated(t *testing.T) {
	// define request body
	jobUpload := ct.JobPostRequest{
		Command: "UPLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
			Data:       []byte(fileContent),
			Permission: Public,
		},
	}
	body, err := json.Marshal(&jobUpload)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	client := &http.Client{}
	// create job
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, "e837ndiefh93h34")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Should return unhautorised %d but returneded %d.\n", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestStorageUploadResourceDuplicated(t *testing.T) {
	// Login the user
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)
	if err != nil {
		t.Fatalf("Unable to marshal request body: %s.\n", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	if fileID == "" {
		t.Fatalf("Resource name cannot be nil.\n")
	}

	// define request body
	jobUpload := ct.JobPostRequest{
		Command: "UPLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
			Data:       []byte(fileContent),
			Permission: Public,
		},
	}
	body, err = json.Marshal(&jobUpload)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to submit resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)
	var jobUploadResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobUploadResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobUploadResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedBody := verifyJobCompletion(t, jobUploadResponse.JobID, session.Token, 30*time.Second)

	// get response body
	var completedUpload ct.JobGetRequest
	err = json.Unmarshal(completedBody, &completedUpload)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}

	// define request body
	jobUpload2 := ct.JobPostRequest{
		Command: "UPLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
			Data:       []byte(fileContent),
			Permission: Public,
		},
	}
	body, err = json.Marshal(&jobUpload2)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	// repeat job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Should return the statues error %d but returned %d.\n", http.StatusInternalServerError, resp.StatusCode)
	}

	// delete resource
	jobDelete := ct.JobPostRequest{
		Command: "DELETE",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
		},
	}
	body, err = json.Marshal(&jobDelete)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}
	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to delete resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)

	var jobDeleteResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobDeleteResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobDeleteResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedDeleteBody := verifyJobCompletion(t, jobDeleteResponse.JobID, session.Token, 30*time.Second)
	// get response body
	var completedDelete ct.JobGetRequest
	err = json.Unmarshal(completedDeleteBody, &completedDelete)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}
}

func TestStorageDeleteJobDuplicated(t *testing.T) {
	// Login the user
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)
	if err != nil {
		t.Fatalf("Unable to marshal request body: %s.\n", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	if fileID == "" {
		t.Fatalf("Resource name cannot be nil.\n")
	}

	// define request body
	jobUpload := ct.JobPostRequest{
		Command: "UPLOAD",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
			Data:       []byte(fileContent),
			Permission: Public,
		},
	}
	body, err = json.Marshal(&jobUpload)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to submit resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)
	var jobUploadResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobUploadResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobUploadResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedBody := verifyJobCompletion(t, jobUploadResponse.JobID, session.Token, 30*time.Second)

	// get response body
	var completedUpload ct.JobGetRequest
	err = json.Unmarshal(completedBody, &completedUpload)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}

	// delete resource
	jobDelete := ct.JobPostRequest{
		Command: "DELETE",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
		},
	}
	body, err = json.Marshal(&jobDelete)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}
	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Unable to delete resource, returned %d but expected %d.\n", resp.StatusCode, http.StatusAccepted)
	}
	respBody, _ = ioutil.ReadAll(resp.Body)

	var jobDeleteResponse ct.JobPostResponse
	err = json.Unmarshal(respBody, &jobDeleteResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if jobDeleteResponse.JobID == "" {
		t.Fatalf("Invalid job id: should not be nil.\n")
	}
	resp.Body.Close()

	// get job completed
	completedDeleteBody := verifyJobCompletion(t, jobDeleteResponse.JobID, session.Token, 30*time.Second)
	// get response body
	var completedDelete ct.JobGetRequest
	err = json.Unmarshal(completedDeleteBody, &completedDelete)
	if err != nil {
		t.Fatalf("Unable to unmarshal data: %s.\n", err.Error())
	}

	// delete resource repeted
	jobDelete2 := ct.JobPostRequest{
		Command: "DELETE",
		Arguments: &ct.CommandArguments{
			ResourceID: fileID,
		},
	}
	body, err = json.Marshal(&jobDelete2)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}
	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/storage/job", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the storage/job request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform job request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Deleting again a resource should produce an error, returned %d but expected %d.\n", resp.StatusCode, http.StatusNotFound)
	}
}
