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

// Third party libs
import (
	"gopkg.in/mgo.v2/bson"
)

// Internal packages
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
)

type Mockdb struct {
	addresses string
	user      string
	password  string
	authDb    string
	// in memory storage
	willsStorage  map[string]will.Will
	emailsStorage map[string]types.Email
}

func NewMockDb(args *ct.DbArgs) *Mockdb {
	return &Mockdb{
		addresses:     ct.ComposeDbAddress(args),
		user:          args.User,
		password:      args.Password,
		authDb:        args.AuthDb,
		willsStorage:  make(map[string]will.Will),
		emailsStorage: make(map[string]types.Email),
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

func (d *Mockdb) RemoveExausted() error {
	for k, v := range d.willsStorage {
		if v.Removable == true {
			err := d.RemoveWill(k)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Mockdb) RemoveWill(id string) error {
	if _, ok := d.willsStorage[id]; !ok {
		return fmt.Errorf("unable to find required %s will", id)
	}
	delete(d.willsStorage, id)
	return nil
}

func (d *Mockdb) SetEmail(email *types.Email) error {
	if email.ObjectID == "" {
		email.ObjectID = bson.NewObjectId()
	}
	d.emailsStorage[string(email.ObjectID)] = *email
	return nil
}

func (d *Mockdb) GetEmails() ([]types.Email, error) {
	result := make([]types.Email, 0)
	for k, v := range d.emailsStorage {
		if v.Sended != true {
			result = append(result, v)
			v.Sended = true
		}
		d.emailsStorage[k] = v
	}
	return result, nil
}

// RemoveSendedEmails remove sended emails while possible.
func (d *Mockdb) RemoveSendedEmails() error {
	for k, v := range d.emailsStorage {
		if v.Sended == true {
			delete(d.emailsStorage, k)
		}
	}
	return nil
}
