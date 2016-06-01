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
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}

	var delta int
	if len(kTestFileContent)%kChunkSize != 0 {
		delta = 1
	}
	if len(chunks.chunks) != (len(kTestFileContent)/kChunkSize)+delta {
		t.Fatalf("Unexpected number of chunks: should have %d but found %d.\n", (len(kTestFileContent)/kChunkSize)+delta, len(chunks.chunks))
	}

	for idx, value := range chunks.chunks {
		if len(value) == 0 {
			t.Fatalf("Unexpected data chunk, lenght should be not nil")
		}
		if len(value) > (kChunkSize + 32 + 4) {
			t.Logf("Chunk too big, should be max %d bytes but found %d at index %d.\n", kChunkSize, len(value), idx)
		}
		key := chunks.chunksKeys[idx]
		if len(key) == 0 {
			t.Fatalf("Key is nil, should not be the case.\n")
		}
	}
}
