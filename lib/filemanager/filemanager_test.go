//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	kMasterKey       = "ThisIsTheMasterKey000123ThisIsTheMasterKey000123"
	kTestFileContent = `This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.
	This is fake data that can
	be used to verify algorithm capabilities, if
	all where fine the whole content would be
	encrypted and stripped.`
	kChunkSize = 500
)

func createTmpFile(content []byte) (string, error) {
	tmpfile, err := ioutil.TempFile("", "3nigm4")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write(content); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}

func TestNewEncryptedChunksNoCompression(t *testing.T) {
	// create tmp file
	filePath, err := createTmpFile([]byte(kTestFileContent))
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(filePath) // clean up

	chunks, err := NewEncryptedChunksFromFile([]byte(kMasterKey), filePath, kChunkSize, false)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	t.Logf("Chunks: %v.\n", chunks)
}
