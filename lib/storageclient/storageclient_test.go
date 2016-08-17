// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

package storageclient

// Std golang dependencies.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

const (
	counterLimit = 10
)

var (
	testFileName   = "/tmp/filename.tmp"
	testFileChunks = [][]byte{
		[]byte("This is a test content used to verify"),
		[]byte("the integrity of chunks in client side"),
		[]byte("upload and download operations."),
	}
	testToken = "a49ef848ab30a7830e09ff923ae93e9f"
)

// Global vars
var (
	mockUploadServer   *httptest.Server
	mockDownloadServer *httptest.Server
	delayCounters      *safeDelayCounters
	mockServiceStorage *serviceStorage
)

func mockUploadHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		delayCounters.initCounter(r.URL.Path)
		// verify value
		if delayCounters.value(r.URL.Path) == counterLimit {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(
				&ct.JobGetRequest{
					Complete: true,
				})
		} else {
			w.WriteHeader(http.StatusAccepted)
			delayCounters.add(r.URL.Path, 1)
		}
	case r.Method == "POST":
		// get message BODY
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.Bytes()
		var createResource ct.JobPostRequest
		err := json.Unmarshal(body, &createResource)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(
				ct.StandardResponse{
					ct.NakResponse,
					"error unmarshaling json",
				})
			return
		}
		mockServiceStorage.mtx.Lock()
		mockServiceStorage.storage[createResource.Arguments.ResourceID] = createResource.Arguments.Data
		mockServiceStorage.mtx.Unlock()

		jobId, err := randomID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(
				ct.StandardResponse{
					ct.NakResponse,
					"error managing request",
				})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(
			&ct.JobPostResponse{
				JobID: jobId,
			})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func mockDownladHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		delayCounters.initCounter(r.URL.Path)
		// verify value
		if delayCounters.value(r.URL.Path) == counterLimit {
			// retrieve resource
			pathComponents := strings.Split(r.URL.Path, "/")
			if len(pathComponents) < 2 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(
					ct.StandardResponse{
						ct.NakResponse,
						"unable to retrieve jobID",
					})
				return
			}
			var jobID string
			if pathComponents[len(pathComponents)-1] != "" {
				jobID = pathComponents[len(pathComponents)-1]
			} else {
				jobID = pathComponents[len(pathComponents)-2]
			}
			mockServiceStorage.mtx.Lock()
			resourceID, ok := mockServiceStorage.jobs[jobID]
			mockServiceStorage.mtx.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(
					ct.StandardResponse{
						ct.NakResponse,
						fmt.Sprintf("job %s not found in storage", jobID),
					})
				return
			}
			mockServiceStorage.mtx.Lock()
			data, ok := mockServiceStorage.storage[resourceID]
			mockServiceStorage.mtx.Unlock()
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(
					ct.StandardResponse{
						ct.NakResponse,
						fmt.Sprintf("unable to find storage resource %s", resourceID),
					})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(
				&ct.JobGetRequest{
					Complete: true,
					Data:     data,
				})
		} else {
			w.WriteHeader(http.StatusAccepted)
			delayCounters.add(r.URL.Path, 1)
		}
	case r.Method == "POST":
		// get message BODY
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.Bytes()
		var retrieveResource ct.JobPostRequest
		err := json.Unmarshal(body, &retrieveResource)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(
				ct.StandardResponse{
					ct.NakResponse,
					"error unmarshaling json",
				})
			return
		}

		jobId, err := randomID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(
				ct.StandardResponse{
					ct.NakResponse,
					"error managing request",
				})
			return
		}

		mockServiceStorage.mtx.Lock()
		mockServiceStorage.jobs[jobId] = retrieveResource.Arguments.ResourceID
		mockServiceStorage.mtx.Unlock()

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(
			&ct.JobPostResponse{
				JobID: jobId,
			})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestMain(m *testing.M) {
	delayCounters = newSafeDelayCounters()
	mockServiceStorage = &serviceStorage{
		storage: make(map[string][]byte),
		jobs:    make(map[string]string),
	}
	// starting up mock servers
	mockUploadServer = httptest.NewServer(http.HandlerFunc(mockUploadHandler))
	defer mockUploadServer.Close()
	mockDownloadServer = httptest.NewServer(http.HandlerFunc(mockDownladHandler))
	defer mockDownloadServer.Close()

	os.Exit(m.Run())
}

var uploadGeneratedFileNames []string

func TestUploadResources(t *testing.T) {
	addr, port := extractAddressAndPort(mockUploadServer.URL, t)
	sc, err := NewStorageClient(addr, port, testToken, 12, 50)
	if err != nil {
		t.Fatalf("Unable to create a new StorageClient instance: %s.\n", err.Error())
	}
	defer sc.Close()

	// manage errrors
	errorCounter := wq.AtomicCounter{}
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-sc.ErrorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()

	fnames, err := sc.SaveChunks(
		testFileName,
		testFileChunks,
		nil,
		nil,
		&fm.Permission{
			Permission: 2,
		})
	if err != nil {
		t.Fatalf("Unable to uplaod files: %s.\n", err.Error())
	}
	if errorCounter.Value() != 0 {
		t.Fatalf("Error counter is not nil, last error: %s.\n", lastError.Error())
	}
	if len(fnames) != len(testFileChunks) {
		t.Fatalf("Unexpected number of resource names: having %d expecting %d.\n", len(fnames), len(testFileChunks))
	}
	t.Logf("Chunks resource IDs: %v.\n", fnames)
	for idx, resourceID := range fnames {
		storageRes, ok := mockServiceStorage.storage[resourceID]
		if !ok {
			t.Fatalf("Resource %s not found.\n", resourceID)
		}
		if bytes.Compare(storageRes, testFileChunks[idx]) != 0 {
			t.Fatalf("Resources at index %d are not equal as expected.\n", idx)
		}
	}
	uploadGeneratedFileNames = fnames
}

func TestDownloadResources(t *testing.T) {
	if uploadGeneratedFileNames == nil {
		t.Fatalf("Tests must be executed in order: previous TestUploadResources should prepare  \"uploadGeneratedFileNames\" var for this test.\n")
	}

	addr, port := extractAddressAndPort(mockDownloadServer.URL, t)
	sc, err := NewStorageClient(addr, port, testToken, 12, 50)
	if err != nil {
		t.Fatalf("Unable to create a new StorageClient instance: %s.\n", err.Error())
	}
	defer sc.Close()

	// manage errrors
	errorCounter := wq.AtomicCounter{}
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-sc.ErrorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()

	chunks, err := sc.RetrieveChunks(testFileName, uploadGeneratedFileNames)
	if err != nil {
		t.Fatalf("Unable to retrieve files: %s.\n", err.Error())
	}
	if errorCounter.Value() != 0 {
		t.Fatalf("Error counter is not nil, last error: %s.\n", lastError.Error())
	}
	if len(chunks) != len(uploadGeneratedFileNames) {
		t.Fatalf("Unexpected number of chunks: having %d expecting %d.\n", len(chunks), len(uploadGeneratedFileNames))
	}
	t.Logf("Chunks slice: %v.\n", chunks)
	for idx, chunk := range chunks {
		if bytes.Compare(chunk, testFileChunks[idx]) != 0 {
			t.Fatalf("Downloaded resources at index %d are not equal as expected %s != %s.\n", idx, string(chunk), string(testFileChunks[idx]))
		}
	}
}
