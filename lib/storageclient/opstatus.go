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

// TotalUnits part of the ProgressStatus
// interface returns the total number of
// units to be processed.
func (p *Progress) TotalUnits() int {
	return p.Total
}

// Done part of the ProgressStatus
// interface returns the number of already
// processed elements.
func (p *Progress) Done() int {
	return p.Progress
}

// Status the status of the required resource operation is
// used to asyncronously return the data passed back by the
// job GET method from APIs.
type Status struct {
	Done bool
	Err  error
	Data []byte
}

// RequestStatus gloabally request related status infos is
// used to maintain, protected in a concurrent environment,
// all resources status and a global progress data (that can
// be used to report the progress status to the UI).
type RequestStatus struct {
	mtx sync.Mutex
	ID  string
	res map[string]Status
	Progress
}

// generateTranscationID generate a randomised tx id to
// avoid conflicts caused, for example, by multiple accesses
// to the same file in a short time range.
func generateTranscationID(id string, t *time.Time) string {
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

// SetStatus set the status for a specified resource id. It sets
// done flag, and a result cd.OpResult structure.
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

// GetStatus retrieve the status of a specified resource id,
// return the status struct and a bool reporting the presence
// of the resource or not (it works as the std golang map primitive).
func (rs *RequestStatus) GetStatus(id string) (*Status, bool) {
	rs.mtx.Lock()
	status, ok := rs.res[id]
	rs.mtx.Unlock()
	if !ok {
		return nil, false
	}
	return &status, true
}

// Completed returns true if all resources composing a request
// have been processed or not, this do not means that no error
// occurred but only that API games are over.
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
