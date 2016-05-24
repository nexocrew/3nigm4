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

func generateChunkRandomKeys(chunksize uint64) ([][]byte, error) {
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

func (e *encryptedChunk) splitDataInChunks(data []byte) error {
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

func NewEncryptedChunk(masterkey []byte, chunkSize uint64, compressed bool) (*encryptedChunk, error) {
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
