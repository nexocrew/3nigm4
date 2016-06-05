//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
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
func createTmpDirAndFiles(content []byte) (string, error) {
	tmpdir, err := ioutil.TempDir("", "3nigm4dir")
	if err != nil {
		return "", err
	}
	txtFile := filepath.Join(tmpdir, "textfile.txt")
	err = ioutil.WriteFile(txtFile, content, 0666)
	if err != nil {
		return "", err
	}
	binFile := filepath.Join(tmpdir, "binfile.bin")
	err = ioutil.WriteFile(binFile, gzipData(content), 0666)
	if err != nil {
		return "", err
	}
	readmeFile := filepath.Join(tmpdir, "README")
	err = ioutil.WriteFile(readmeFile, []byte("This is a test repo dir."), 0666)
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(tmpdir, "data")
	err = os.Mkdir(dataDir, 0777)
	if err != nil {
		return "", err
	}
	secondTxtFile := filepath.Join(dataDir, "txtfile.txt")
	err = ioutil.WriteFile(secondTxtFile, content, 0666)
	if err != nil {
		return "", err
	}
	return tmpdir, nil
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

	recomposedData := make([]byte, 0)
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
		decrypted, err := crypto3n.AesDecrypt(key, value, crypto3n.CBC)
		if err != nil {
			t.Fatalf("Unexpected error decrypting chunk: %s.\n", err.Error())
		}
		if len(decrypted) == 0 {
			t.Fatalf("Decrypted chunk is nil: having %d len.\n", len(decrypted))
		}
		if len(decrypted) > kChunkSize {
			t.Fatalf("Decrypted chunk is too big: having %d expecting max %d.\n", len(decrypted), kChunkSize)
		}
		recomposedData = append(recomposedData, decrypted...)
	}
	if bytes.Compare(recomposedData, []byte(kTestFileContent)) != 0 {
		t.Fatalf("Recomposed data do not match with original data.\n")
	}

	// test restoring
	tmpfile, err := ioutil.TempFile("", "3nigm4")
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	//copiedChunks := copyEncryptedChunks(chunks)
	err = chunks.FileFromEncryptedChunks(tmpfile.Name())
	if err != nil {
		t.Fatalf("Unable to restore file: %s.\n", err.Error())
	}
}

func TestNewEncryptedChunksWithCompression(t *testing.T) {
	// create tmp file
	filePath, err := createTmpFile([]byte(kTestFileContent))
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(filePath) // clean up

	chunks, err := NewEncryptedChunksFromFile([]byte(kMasterKey), filePath, kChunkSize, true)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}

	// test restoring
	tmpfile, err := ioutil.TempFile("", "3nigm4")
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	//copiedChunks := copyEncryptedChunks(chunks)
	err = chunks.FileFromEncryptedChunks(tmpfile.Name())
	if err != nil {
		t.Fatalf("Unable to restore file: %s.\n", err.Error())
	}
}

func TestNewEncryptedChunksDirectoryWithCompression(t *testing.T) {
	// create tmp file
	dirPath, err := createTmpDirAndFiles([]byte(kTestFileContent))
	if err != nil {
		t.Fatalf("Unable to create tmp directory: %s.\n", err.Error())
	}

	chunks, err := NewEncryptedChunksFromFile([]byte(kMasterKey), dirPath, kChunkSize, true)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}
	os.RemoveAll(dirPath) // clean up

	//copiedChunks := copyEncryptedChunks(chunks)
	err = chunks.FileFromEncryptedChunks("/tmp")
	if err != nil {
		t.Fatalf("Unable to restore file: %s.\n", err.Error())
	}
	defer os.RemoveAll(dirPath)

	// check extracted files
	readmeFile := filepath.Join(dirPath, "README")
	finfo, err := os.Stat(readmeFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", readmeFile, err.Error())
	}
	if finfo.Size() != 24 {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", readmeFile, finfo.Size(), 24)
	}
	txtFile := filepath.Join(dirPath, "textfile.txt")
	finfo, err = os.Stat(txtFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", txtFile, err.Error())
	}
	if finfo.Size() != int64(len(kTestFileContent)) {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", txtFile, finfo.Size(), len(kTestFileContent))
	}
	binFile := filepath.Join(dirPath, "binfile.bin")
	finfo, err = os.Stat(binFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", binFile, err.Error())
	}
	dataDir := filepath.Join(dirPath, "data")
	finfo, err = os.Stat(dataDir)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", dataDir, err.Error())
	}
	if finfo.IsDir() != true {
		t.Fatalf("Unexpected stat type, should be a directory.\n")
	}
	txtSecondFile := filepath.Join(dirPath, "data", "txtfile.txt")
	finfo, err = os.Stat(txtSecondFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", txtSecondFile, err.Error())
	}
	if finfo.Size() != int64(len(kTestFileContent)) {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", txtSecondFile, finfo.Size(), len(kTestFileContent))
	}

}
