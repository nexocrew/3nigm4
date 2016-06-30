//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package s3backend

import (
	"testing"
	"time"
)

import (
	"github.com/nexocrew/3nigm4/lib/itm"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

const (
	kFileContent = `Test this content for file usage,
		should be used to test upload functions to the
		S£ instance.`
	kFileId = "testfile"
)

func TestS3UploadInterface(t *testing.T) {
	s3, err := NewS3BackendSession(itm.S().S3Endpoint(),
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
	var lastError error
	go func() {
		for {
			select {
			case err := <-s3.ErrorChan:
				errorCounter.Add(1)
				lastError = err
				t.Logf("%v", err)
			case response := <-s3.SendedChan:
				processedCounter.Add(1)
				t.Logf("%v", response)
			}
		}
	}()

	s3.Upload(itm.S().S3Bucket(), kFileId, []byte(kFileContent), nil)
	defer s3.Delete(itm.S().S3Bucket(), kFileId)

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
		if processedCounter.Value() == 1 {
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
