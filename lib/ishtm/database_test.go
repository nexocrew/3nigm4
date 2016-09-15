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
	willsStorage map[string]Will
}

func newMockDb(args *DbArgs) *mockdb {
	return &mockdb{
		addresses:    composeDbAddress(args),
		user:         args.User,
		password:     args.Password,
		authDb:       args.AuthDb,
		willsStorage: make(map[string]Will),
	}
}

func (d *mockdb) Copy() Database {
	return d
}

func (d *mockdb) Close() {
}

func (d *mockdb) GetWills(owner string) ([]Will, error) {
	result := make([]Will, 0)
	for _, value := range d.willsStorage {
		if value.Owner.Name == owner {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("unable to find required owner %s", owner)
	}
	return result, nil
}

func (d *mockdb) GetWill(id string) (*Will, error) {
	will, ok := d.willsStorage[id]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s will", id)
	}
	return &will, nil
}

func (d *mockdb) SetWill(will *Will) error {
	_, ok := d.willsStorage[will.ID]
	if ok {
		return fmt.Errorf("will %s already exist in the db", will.ID)
	}
	d.willsStorage[will.ID] = *will
	return nil
}

func (d *mockdb) RemoveUser(id string) error {
	if _, ok := d.willsStorage[id]; !ok {
		return fmt.Errorf("unable to find required %s will", id)
	}
	delete(d.willsStorage, id)
	return nil
}

func (d *mockdb) GetInDelivery(actual time.Time) ([]Will, error) {
	result := make([]Will, 0)
	for _, value := range d.willsStorage {
		if value.TimeToDelivery.Sub(actual.UTC()) < 0 {
			result = append(result, value)
		}
	}
	return result, nil
}
