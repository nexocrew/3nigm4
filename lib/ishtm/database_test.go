//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//
// This mock database is used for tests purposes, should
// never be used in production environment. It's not
// concurrency safe and do not implement any performance
// optimisation logic.
//

package ishtm

// Golang std libs
import (
	"fmt"
	"time"
)

type mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	jobsStorage map[string]Job
}

func newMockDb(args *DbArgs) *mockdb {
	return &mockdb{
		addresses:   composeDbAddress(args),
		user:        args.User,
		password:    args.Password,
		authDb:      args.AuthDb,
		jobsStorage: make(map[string]Job),
	}
}

func (d *mockdb) Copy() Database {
	return d
}

func (d *mockdb) Close() {
}

func (d *mockdb) GetJobs(owner string) ([]Job, error) {
	result := make([]Job, 0)
	for _, value := range d.jobsStorage {
		if value.Owner.Name == owner {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("unable to find required owner %s", owner)
	}
	return result, nil
}

func (d *mockdb) GetJob(id string) (*Job, error) {
	job, ok := d.jobsStorage[id]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s job", id)
	}
	return &job, nil
}

func (d *mockdb) SetJob(job *Job) error {
	_, ok := d.jobsStorage[job.ID]
	if ok {
		return fmt.Errorf("job %s already exist in the db", job.ID)
	}
	d.jobsStorage[job.ID] = *job
	return nil
}

func (d *mockdb) RemoveUser(id string) error {
	if _, ok := d.jobsStorage[id]; !ok {
		return fmt.Errorf("unable to find required %s job", id)
	}
	delete(d.jobsStorage, id)
	return nil
}

func (d *mockdb) GetInDelivery(actual time.Time) ([]Job, error) {
	result := make([]Job, 0)
	for _, value := range d.jobsStorage {
		if value.TimeToDelivery.Sub(actual.UTC()) < 0 {
			result = append(result, value)
		}
	}
	return result, nil
}
