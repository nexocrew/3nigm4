//
// 3nigm4 ishtmmocks package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//
// This mock database is used for tests purposes, should
// never be used in production environment. It's not
// concurrency safe and do not implement any performance
// optimisation logic.
//

// Package ishtmmocks implements unit-tests mocks for the
// ishtm service.
package ishtmmocks

// Golang std libs
import (
	"fmt"
	"time"
)

// Internal packages
import (
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
)

type Mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	willsStorage map[string]will.Will
}

func NewMockDb(args *ct.DbArgs) *Mockdb {
	return &Mockdb{
		addresses:    ct.ComposeDbAddress(args),
		user:         args.User,
		password:     args.Password,
		authDb:       args.AuthDb,
		willsStorage: make(map[string]will.Will),
	}
}

func (d *Mockdb) Copy() ct.Database {
	return d
}

func (d *Mockdb) Close() {
}

func (d *Mockdb) GetWills(owner string) ([]will.Will, error) {
	result := make([]will.Will, 0)
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

func (d *Mockdb) GetWill(id string) (*will.Will, error) {
	will, ok := d.willsStorage[id]
	if !ok {
		return nil, fmt.Errorf("unable to find the required %s will", id)
	}
	return &will, nil
}

func (d *Mockdb) SetWill(will *will.Will) error {
	_, ok := d.willsStorage[will.ID]
	if ok {
		return fmt.Errorf("will %s already exist in the db", will.ID)
	}
	d.willsStorage[will.ID] = *will
	return nil
}

func (d *Mockdb) GetInDelivery(actual time.Time) ([]will.Will, error) {
	result := make([]will.Will, 0)
	for _, value := range d.willsStorage {
		if value.TimeToDelivery.Sub(actual.UTC()) < 0 {
			result = append(result, value)
		}
	}
	return result, nil
}

func (d *Mockdb) RemoveWill(id string) error {
	if _, ok := d.willsStorage[id]; !ok {
		return fmt.Errorf("unable to find required %s will", id)
	}
	delete(d.willsStorage, id)
	return nil
}
