// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

package storageclient

// Std golang dependencies.
import (
	"fmt"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

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

func (s *StorageClient) updateDownloadRequestStatus(downloaded ct.OpResult) {
	// upload request
	value, ok := s.requests[downloaded.RequestID]
	if !ok {
		s.ErrorChan <- fmt.Errorf("unable to find request status manager for %s", downloaded.RequestID)
		return
	}
	err := value.SetStatus(downloaded.ID, true, &downloaded)
	if err != nil {
		s.ErrorChan <- err
		return
	}
}

func (s *StorageClient) updateDeleteRequestStatus(deleted ct.OpResult) {
	// upload request
	value, ok := s.requests[deleted.RequestID]
	if !ok {
		s.ErrorChan <- fmt.Errorf("unable to find request status manager for %s", deleted.RequestID)
		return
	}
	err := value.SetStatus(deleted.ID, true, &deleted)
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
				go s.updateDownloadRequestStatus(downloaded)
			}
		case deleted, deletedcOk := <-s.deletedChan:
			if !deletedcOk {
				deletedcClosed = true
			} else {
				go s.updateDeleteRequestStatus(deleted)
			}
		}
	}
}
