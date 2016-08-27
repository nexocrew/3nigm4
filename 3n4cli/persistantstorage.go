//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"time"
)

// pss (persistant storage singleton) is a global var used
// to maintain the singleton instance of the storage.
var pss *storage

const (
	storageFileName = ".storage" // file name where storage info will be saved (under app root dir).
)

// storage struct is used to structure the persistant
// store kept on the fs by the app to maintain it's data between
// commands invokations.
type storage struct {
	Token     string    `json:"token" xml:"token"`
	LastLogin time.Time `json:"lastlogin" xml:"lastlogin"`
}

// storageFilePath returns the file path of the storage
// file.
func storageFilePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	filePath := path.Join(usr.HomeDir, rootAppFolder, storageFileName)
	return filePath, nil
}

// newPersistentStorage returns the singleton instance or a newly
// created instance (even if something went wrong).
func newPersistentStorage() *storage {
	var ps *storage = nil
	var err error
	if pss == nil {
		ps, err = persistentStorageRead()
		if err != nil {
			log.WarningLog("Persistant storage not found, creating it!\n")
			ps = &storage{}
		}
	} else {
		ps = pss
	}
	return ps
}

// persistentStorageRead reads persistent storage from the fs.
func persistentStorageRead() (*storage, error) {
	filePath, err := storageFilePath()
	if err != nil {
		return nil, err
	}
	binary, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read persistant storage file cause %s", err.Error())
	}
	var ps storage
	err = xml.Unmarshal(binary, &ps)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal xml data cause %s", err.Error())
	}
	return &ps, nil
}

// save save on disk mutated structure (always called on
// program exit).
func (ps *storage) save() error {
	binary, err := xml.Marshal(ps)
	if err != nil {
		return fmt.Errorf("unable to marshal persistant storage cause %s", err.Error())
	}
	filePath, err := storageFilePath()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, binary, 0600)
	if err != nil {
		return fmt.Errorf("unable to save persistant storage to fs cause %s", err.Error())
	}
	return nil
}

// remove delete the persistant storage file from the fs.
func (ps *storage) remove() error {
	filePath, err := storageFilePath()
	if err != nil {
		return err
	}
	return os.Remove(filePath)
}

// invalidateSessionToken invalidate session token.
func (ps *storage) invalidateSessionToken() {
	ps.Token = ""
}

// refreshLastLogin refresh last login date.
func (ps *storage) refreshLastLogin() {
	ps.LastLogin = time.Now()
}
