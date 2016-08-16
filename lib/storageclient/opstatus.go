// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

package storageclient

// Std golang dependencies.
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Progress is used to maintain the status of various files
// update, it report the initial number of files and as soon
// as upload successfully terminates, the number of uploaded
// parts.
type Progress struct {
	Total    int
	Progress int
	Errors   int
}

type Status struct {
	Done bool
	Err  error
	Data []byte
}

type RequestStatus struct {
	mtx sync.Mutex
	ID  string
	res map[string]Status
	Progress
}

// generateTranscationId generate a randomised tx id to
// avoid conflicts caused, for example, by multiple accesses
// to the same file in a short time range.
func generateTranscationId(id string, t *time.Time) string {
	var raw []byte
	randoms, _ := ct.RandomBytesForLen(16)
	raw = append(raw, randoms...)
	composed := fmt.Sprintf("%s.%d.%d",
		id,
		t.Unix(),
		t.UnixNano())
	raw = append(raw, []byte(composed)...)
	checksum := sha256.Sum256(raw)
	return hex.EncodeToString(checksum[:])
}

// NewRequestStatus creates a new request status, having as
// pre-requisite that all the ids are different (should always
// be the case in 3nigm4 scenario).
func NewRequestStatus(reqID string, count int) *RequestStatus {
	return &RequestStatus{
		ID:  reqID,
		res: make(map[string]Status),
		Progress: Progress{
			Total: count,
		},
	}
}

func (rs *RequestStatus) SetStatus(id string, done bool, result *ct.OpResult) error {
	rs.mtx.Lock()
	status := Status{
		Done: done,
	}
	if result != nil {
		status.Data = result.Data
		status.Err = result.Error
		if result.Error != nil {
			rs.Progress.Errors++
		}
	}
	if done == true {
		rs.Progress.Progress++
	}
	rs.res[id] = status
	rs.mtx.Unlock()
	return nil
}

func (rs *RequestStatus) GetStatus(id string) (*Status, bool) {
	rs.mtx.Lock()
	status, ok := rs.res[id]
	rs.mtx.Unlock()
	if !ok {
		return nil, false
	}
	return &status, true
}

func (rs *RequestStatus) Completed() bool {
	var incomplete int
	rs.mtx.Lock()
	for _, value := range rs.res {
		if value.Done == false {
			incomplete++
		}
	}
	rs.mtx.Unlock()
	if incomplete != 0 {
		return false
	}
	return true
}
