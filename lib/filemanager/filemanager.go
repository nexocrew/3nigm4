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
	"math"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

type Metadata struct {
	fileName string
	size     int64
}

type EncryptedDataChunkManager interface {
	SetBytes([]byte) ([]byte, [][]byte, error) // returns a byte blob and an array of chunk specific keys.
	GetBytes([][]byte) ([]byte, error)
	SetChunkSize(int64) error
	GetChunkSize() int64
	SetIsCompressed(bool)
	GetIsCompressed() bool
	GetFileMetadata() *Metadata
	SetFileMetadata() error
}

type encryptedChunk struct {
	masterKey  []byte
	compressed bool
	chunkSize  int64
	chunks     [][]byte
	chunksKeys [][]byte
	metadata   Metadata
}

const (
	minKeyLen    = 32
	minChunkSize = 32
	chunkKeySize = 32
)

func generateChunkRandomKeys(chunksize int) ([][]byte, error) {
	chunkKeys := make([][]byte, chunksize)
	for idx := 0; idx < chunksize; idx++ {
		buf := make([]byte, chunkKeySize)
		_, err := rand.Read(buf)
		if err != nil {
			return nil, err
		}
		chunkKeys[idx] = buf
	}
	return chunkKeys, nil
}

func (e *encryptedChunk) splitDataInChunks(data []byte) error {
	// check if chunk size is bigger than file size
	var totalPartsCount uint64
	if len(data) <= e.chunkSize {
		totalPartsCount = 1
	} else {
		totalPartsCount = uint64(math.Ceil(float64(len(data)) / float64(e.chunkSize)))
	}

	for idx := uint64(0); idx < totalPartsCount; idx++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		// split data blob
	}
}

func NewEncryptedChunk(masterkey []byte, chunkSize int64, compressed bool) (*encryptedChunk, error) {
	if len(masterkey) < minKeyLen {
		return nil, fmt.Errorf("unable to create an encrypted chunk, key is too short: having %d expecting %d", minKeyLen, len(masterkey))
	}
	if chunkSize < minChunkSize {
		return nil, fmt.Errorf("required chunk size is too small: should be major than %d", minChunkSize)
	}
	randomKeys, err := generateChunkRandomKeys(chunksize)
	if err != nil {
		return nil, err
	}
	return &encryptedChunk{
		masterkey:  masterkey,
		compressed: compressed,
		chunkSize:  chunkSize,
		chunksKeys: randomKeys,
	}
}
