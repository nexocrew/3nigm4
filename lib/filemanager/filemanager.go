//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

type Metadata struct {
	FileName string
	Size     int64
	ModTime  time.Time
	IsDir    bool
}

type EncryptedChunks struct {
	masterKey  []byte
	compressed bool
	chunkSize  uint64
	chunks     [][]byte
	chunksKeys [][]byte
	metadata   Metadata
}

const (
	minKeyLen    = 32
	minChunkSize = 32
	chunkKeySize = 32
)

func generateChunksRandomKeys(chunksize uint64) ([][]byte, error) {
	chunkKeys := make([][]byte, chunksize)
	for idx := uint64(0); idx < chunksize; idx++ {
		buf := make([]byte, chunkKeySize)
		_, err := rand.Read(buf)
		if err != nil {
			return nil, err
		}
		chunkKeys[idx] = buf
	}
	return chunkKeys, nil
}

func (e *EncryptedChunks) splitDataInChunks(data []byte) error {
	// check if chunk size is bigger than file size
	var totalPartsCount uint64
	if len(data) <= e.chunkSize {
		totalPartsCount = 1
	} else {
		totalPartsCount = uint64(math.Ceil(float64(len(data)) / float64(e.chunkSize)))
	}

	// before starting check for keys number
	if len(e.chunksKeys) != totalPartsCount {
		return fmt.Errorf("unexpected number of keys, having %d parts and %d keys", totalPartsCount, len(e.chunksKeys))
	}

	// init chunks structure
	e.chunks = make([][]byte, totalPartsCount)

	for idx := uint64(0); idx < totalPartsCount; idx++ {
		processedLen := len(data) - int64(idx*e.chunkSize)
		partSize := uint64(math.Min(float64(e.chunkSize), float64(processedLen)))
		partBuffer := make([]byte, 0)
		// copy data blob
		partBuffer = append(partBuffer, data[processedLen:processedLen+partSize]...)
		// generate random salt of len 8 bytes
		salt := make([]byte, 8)
		_, err := rand.Read(salt)
		if err != nil {
			return err
		}
		// encrypt using associated key
		encryptedChunk, err := crypto3n.AesEncrypt(e.chunksKeys[idx], salt, partBuffer, crypto3n.CBC)
		if err != nil {
			return err
		}
		// assign to internal buffer
		e.chunks[idx] = encryptedChunk
	}
	return nil
}

func compressAndSaveToTempFile(dirpath string) (string, error) {
	/*
		content := []byte("temporary file's content")
		tmpfile, err := ioutil.TempFile("", "example")
		if err != nil {
			log.Fatal(err)
		}

		defer os.Remove(tmpfile.Name()) // clean up

		if _, err := tmpfile.Write(content); err != nil {
			log.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			log.Fatal(err)
		}
	*/
}

func NewEncryptedChunksFromFile(masterkey []byte, filepath string, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	// get infos from file
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	var finalFilePath string
	if fileInfo.IsDir() == true {
		// create a tmp tgz archive and operate on it
	} else {
		// assign final path to actual one
		finalFilePath = filepath
	}

	// create chunk struct
	chunk, err := newEncryptedChunks(masterkey, chunkSize, compressed)
	if err != nil {
		return nil, err
	}
	// define metadata
	chunk.metadata.FileName = fileInfo.Name()
	chunk.metadata.Size = fileInfo.Size()
	chunk.metadata.IsDir = fileInfo.IsDir()
	chunk.metadata.ModTime = fileInfo.ModTime()

	// read argument passed file
	data, err := ioutil.ReadFile(finalFilePath)
	if err != nil {
		return nil, err
	}
	err = chunk.splitDataInChunks(data)
	if err != nil {
		return nil, err
	}

	return chunk, nil
}

func newEncryptedChunks(masterkey []byte, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	if len(masterkey) < minKeyLen {
		return nil, fmt.Errorf("unable to create an encrypted chunk, key is too short: having %d expecting %d", minKeyLen, len(masterkey))
	}
	if chunkSize < minChunkSize {
		return nil, fmt.Errorf("required chunk size is too small: should be major than %d", minChunkSize)
	}
	randomKeys, err := generateChunksRandomKeys(chunksize)
	if err != nil {
		return nil, err
	}
	return &EncryptedChunks{
		masterkey:  masterkey,
		compressed: compressed,
		chunkSize:  chunkSize,
		chunksKeys: randomKeys,
	}
}
