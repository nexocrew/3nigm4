//
// 3nigm4 filemanager package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

package filemanager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

func createCustomSizedTmpFile(content []byte, size uint64) (string, error) {
	tmpfile, err := ioutil.TempFile("", "3nigm4")
	if err != nil {
		return "", err
	}
	var lenght uint64
	for lenght <= size {
		written, err := tmpfile.Write(content)
		if err != nil {
			return "", err
		}
		lenght += uint64(written)
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

	chunks, err := NewEncryptedChunks(nil, filePath, kChunkSize, false)
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

	var recomposedData []byte
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
	err = chunks.GetFile(tmpfile.Name())
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

	chunks, err := NewEncryptedChunks(nil, filePath, kChunkSize, true)
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
	err = chunks.GetFile(tmpfile.Name())
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

	chunks, err := NewEncryptedChunks(nil, dirPath, kChunkSize, true)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}
	os.RemoveAll(dirPath) // clean up

	//copiedChunks := copyEncryptedChunks(chunks)
	tmpdir, err := ioutil.TempDir("", "3nigm4dir")
	if err != nil {
		t.Fatalf("Unable to create tmp dir: %s.\n", err.Error())
	}
	err = chunks.GetFile(tmpdir)
	if err != nil {
		t.Fatalf("Unable to restore file: %s.\n", err.Error())
	}
	defer os.RemoveAll(tmpdir)

	// check extracted files
	tmpdir = filepath.Join(tmpdir, chunks.metadata.FileName)
	readmeFile := filepath.Join(tmpdir, "README")
	finfo, err := os.Stat(readmeFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", readmeFile, err.Error())
	}
	if finfo.Size() != 24 {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", readmeFile, finfo.Size(), 24)
	}
	txtFile := filepath.Join(tmpdir, "textfile.txt")
	finfo, err = os.Stat(txtFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", txtFile, err.Error())
	}
	if finfo.Size() != int64(len(kTestFileContent)) {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", txtFile, finfo.Size(), len(kTestFileContent))
	}
	binFile := filepath.Join(tmpdir, "binfile.bin")
	finfo, err = os.Stat(binFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", binFile, err.Error())
	}
	dataDir := filepath.Join(tmpdir, "data")
	finfo, err = os.Stat(dataDir)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", dataDir, err.Error())
	}
	if finfo.IsDir() != true {
		t.Fatalf("Unexpected stat type, should be a directory.\n")
	}
	txtSecondFile := filepath.Join(tmpdir, "data", "txtfile.txt")
	finfo, err = os.Stat(txtSecondFile)
	if err != nil {
		t.Fatalf("Unable to stat %s cause %s.\n", txtSecondFile, err.Error())
	}
	if finfo.Size() != int64(len(kTestFileContent)) {
		t.Fatalf("Unexpected %s file size: having %d expecting %d.\n", txtSecondFile, finfo.Size(), len(kTestFileContent))
	}

}

func TestDataSaverLogics(t *testing.T) {
	// create tmp file
	filePath, err := createCustomSizedTmpFile([]byte(kTestFileContent), 500000)
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(filePath) // clean up

	chunks, err := NewEncryptedChunks(nil, filePath, kChunkSize, false)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}

	// create tmp destination path
	tmpdir, err := ioutil.TempDir("", "datasaver")
	if err != nil {
		t.Fatalf("Unable to define tmp dir: %s.\n", err.Error())
	}
	ds, err := NewLocalDataSaver(tmpdir)
	if err != nil {
		t.Fatalf("Unable to create a new data saver: %s.\n", err.Error())
	}
	defer ds.Cleanup(nil)

	// do it!
	reference, err := chunks.SaveChunks(ds, nil, nil)
	if err != nil {
		t.Fatalf("Unable to save chunks using data saver: %s.\n", err.Error())
	}
	// check files existance
	for _, file := range reference.ChunksPaths {
		path := fmt.Sprintf("%s/%s", tmpdir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Unable to find file %s: %s.\n", path, err.Error())
		}
		if info.Size() > (kChunkSize+32+4) ||
			info.Size() == 0 {
			t.Fatalf("Unexpected file size: %d should be < %d and != 0.\n", info.Size(), (kChunkSize + 32 + 4))
		}
	}

	// recompose it
	recomposedChunks, err := LoadChunks(ds, reference, nil)
	if err != nil {
		t.Fatalf("Unable to load chunks: %s.\n", err.Error())
	}
	if reflect.DeepEqual(recomposedChunks, chunks) != true {
		t.Fatalf("Expected same encrypted chunks structure.\n")
	}

	// verify final data
	recomposed, err := recomposedChunks.composeOriginalData()
	if err != nil {
		t.Fatalf("Unable to recompose data: %s.\n", err.Error())
	}
	original, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Unable to read file: %s.\n", err.Error())
	}
	if bytes.Compare(recomposed, original) != 0 {
		t.Fatalf("Recomposed data do not match original data.\n")
	}
}

func TestDataSaverLogicsWithPassword(t *testing.T) {
	// create tmp file
	filePath, err := createCustomSizedTmpFile([]byte(kTestFileContent), 500000)
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(filePath) // clean up

	rawKey := []byte("testkey0001")
	chunks, err := NewEncryptedChunks(rawKey, filePath, kChunkSize, false)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}

	// create tmp destination path
	tmpdir, err := ioutil.TempDir("", "datasaver")
	if err != nil {
		t.Fatalf("Unable to define tmp dir: %s.\n", err.Error())
	}
	ds, err := NewLocalDataSaver(tmpdir)
	if err != nil {
		t.Fatalf("Unable to create a new data saver: %s.\n", err.Error())
	}
	defer ds.Cleanup(nil)

	// do it!
	reference, err := chunks.SaveChunks(ds, nil, nil)
	if err != nil {
		t.Fatalf("Unable to save chunks using data saver: %s.\n", err.Error())
	}
	// check files existance
	for _, file := range reference.ChunksPaths {
		path := fmt.Sprintf("%s/%s", tmpdir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Unable to find file %s: %s.\n", path, err.Error())
		}
		if info.Size() > (kChunkSize+32+4) ||
			info.Size() == 0 {
			t.Fatalf("Unexpected file size: %d should be < %d and != 0.\n", info.Size(), (kChunkSize + 32 + 4))
		}
	}

	// recompose it
	recomposedChunks, err := LoadChunks(ds, reference, rawKey)
	if err != nil {
		t.Fatalf("Unable to load chunks: %s.\n", err.Error())
	}
	if reflect.DeepEqual(recomposedChunks, chunks) != true {
		t.Fatalf("Expected same encrypted chunks structure.\n")
	}

	// verify final data
	recomposed, err := recomposedChunks.composeOriginalData()
	if err != nil {
		t.Fatalf("Unable to recompose data: %s.\n", err.Error())
	}
	original, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Unable to read file: %s.\n", err.Error())
	}
	if bytes.Compare(recomposed, original) != 0 {
		t.Fatalf("Recomposed data do not match original data.\n")
	}
}

func TestDataSaverLogicsWithWrongPassword(t *testing.T) {
	// create tmp file
	filePath, err := createCustomSizedTmpFile([]byte(kTestFileContent), 500000)
	if err != nil {
		t.Fatalf("Unable to create tmp file: %s.\n", err.Error())
	}
	defer os.Remove(filePath) // clean up

	rawKey := []byte("testkey0001")
	chunks, err := NewEncryptedChunks(rawKey, filePath, kChunkSize, false)
	if err != nil {
		t.Fatalf("Unable to create chunks: %s.\n", err.Error())
	}
	if len(chunks.chunksKeys) != len(chunks.chunks) {
		t.Fatalf("Unexpected slice sizes: having %d expecting %d.\n", len(chunks.chunks), len(chunks.chunksKeys))
	}

	// create tmp destination path
	tmpdir, err := ioutil.TempDir("", "datasaver")
	if err != nil {
		t.Fatalf("Unable to define tmp dir: %s.\n", err.Error())
	}
	ds, err := NewLocalDataSaver(tmpdir)
	if err != nil {
		t.Fatalf("Unable to create a new data saver: %s.\n", err.Error())
	}
	defer ds.Cleanup(nil)

	// do it!
	reference, err := chunks.SaveChunks(ds, nil, nil)
	if err != nil {
		t.Fatalf("Unable to save chunks using data saver: %s.\n", err.Error())
	}
	// check files existance
	for _, file := range reference.ChunksPaths {
		path := fmt.Sprintf("%s/%s", tmpdir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Unable to find file %s: %s.\n", path, err.Error())
		}
		if info.Size() > (kChunkSize+32+4) ||
			info.Size() == 0 {
			t.Fatalf("Unexpected file size: %d should be < %d and != 0.\n", info.Size(), (kChunkSize + 32 + 4))
		}
	}

	// recompose it
	recomposedChunks, err := LoadChunks(ds, reference, nil)
	if err != nil {
		t.Fatalf("Unable to load chunks: %s.\n", err.Error())
	}
	if reflect.DeepEqual(recomposedChunks, chunks) == true {
		t.Fatalf("These structures should not match, the raw key is different.\n")
	}

	// verify final data
	_, err = recomposedChunks.composeOriginalData()
	if err == nil {
		t.Fatalf("Expected an error while decrypting with wrong key.\n")
	}
}
