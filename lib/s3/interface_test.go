//
// 3nigm4 s3backend package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package s3backend

import (
	"testing"
	"time"
)

import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/itm"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

const (
	fileContent = `Test this content for file usage,
		should be used to test upload functions to the
		S3 instance.`
	fileID = "testfile"
)

func TestS3UploadInterface(t *testing.T) {
	s3, err := NewSession(itm.S().S3Endpoint(),
		itm.S().S3Region(),
		itm.S().S3Id(),
		itm.S().S3Secret(),
		itm.S().S3Token(),
		24, 200, true)
	if err != nil {
		t.Fatalf("Unable to create a valid S3 session: %s.\n", err.Error())
	}

	// create resonse listening routine
	errorCounter := wq.AtomicCounter{}
	var uploaded []ct.OpResult
	var lastError error
	go func() {
		for {
			select {
			case err := <-s3.ErrorChan:
				errorCounter.Add(1)
				lastError = err
				t.Logf("%v", err)
			case rUploaded := <-s3.UploadedChan:
				uploaded = append(uploaded, rUploaded)
			case dataDownloaded := <-s3.DownloadedChan:
				t.Logf("%v", dataDownloaded)
			}
		}
	}()

	// upload data
	s3.Upload(itm.S().S3Bucket(), fileID, fileID, []byte(fileContent), nil)

	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (wg waits only for the chan
	// injection).
	ticker := time.Tick(7 * time.Second)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	for {
		if len(uploaded) == 1 {
			if uploaded[0].ID != fileID {
				t.Fatalf("Unexpected file id having %s expecting %s.\n", uploaded[0].ID, fileID)
			}
			if uploaded[0].Error != nil {
				t.Fatalf("Error must be nil but found a valid error.\n")
			}
			break
		}
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Time out reached.\n")
		}
		if errorCounter.Value() != 0 {
			t.Fatalf("Founded an error while uploading the file, %s.\n", lastError.Error())
		}

		time.Sleep(3 * time.Millisecond)
	}
}

func TestS3DownloadInterface(t *testing.T) {
	s3, err := NewSession(itm.S().S3Endpoint(),
		itm.S().S3Region(),
		itm.S().S3Id(),
		itm.S().S3Secret(),
		itm.S().S3Token(),
		24, 200, true)
	if err != nil {
		t.Fatalf("Unable to create a valid S3 session: %s.\n", err.Error())
	}

	// create resonse listening routine
	errorCounter := wq.AtomicCounter{}
	processedCounter := wq.AtomicCounter{}
	var downloaded []ct.OpResult
	var lastError error
	go func() {
		for {
			select {
			case err := <-s3.ErrorChan:
				errorCounter.Add(1)
				lastError = err
				t.Logf("%v", err)
			case rUploaded := <-s3.UploadedChan:
				processedCounter.Add(1)
				t.Logf("%v", rUploaded)
			case dataDownloaded := <-s3.DownloadedChan:
				downloaded = append(downloaded, dataDownloaded)
			}
		}
	}()

	// download data
	s3.Download(itm.S().S3Bucket(), fileID, fileID)

	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (wg waits only for the chan
	// injection).
	ticker := time.Tick(7 * time.Second)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	for {
		if len(downloaded) == 1 {
			if downloaded[0].RequestID != fileID {
				t.Fatalf("Unexpected file id having %s expecting %s.\n", downloaded[0].RequestID, fileID)
			}
			if len(downloaded[0].Data) != len([]byte(fileContent)) {
				t.Fatalf("Unexpected file size, having %d expecting %d.\n", len(downloaded[0].Data), len([]byte(fileContent)))
			}
			break
		}
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Time out reached.\n")
		}
		if errorCounter.Value() != 0 {
			t.Fatalf("Founded an error while uploading the file, %s.\n", lastError.Error())
		}

		time.Sleep(3 * time.Millisecond)
	}
}
