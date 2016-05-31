//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import (
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
	if len(data) <= int(e.chunkSize) {
		totalPartsCount = 1
	} else {
		totalPartsCount = uint64(math.Ceil(float64(len(data)) / float64(e.chunkSize)))
	}

	// generate random keys for chunks
	var err error
	e.chunksKeys, err = generateChunksRandomKeys(totalPartsCount)
	if err != nil {
		return err
	}

	// before starting check for keys number
	if len(e.chunksKeys) != int(totalPartsCount) {
		return fmt.Errorf("unexpected number of keys, having %d parts and %d keys", totalPartsCount, len(e.chunksKeys))
	}

	// init chunks structure
	e.chunks = make([][]byte, totalPartsCount)

	for idx := uint64(0); idx < totalPartsCount; idx++ {
		processedLen := int64(len(data)) - int64(idx*e.chunkSize)
		partSize := uint64(math.Min(float64(e.chunkSize), float64(processedLen)))
		partBuffer := make([]byte, 0)
		// copy data blob
		partBuffer = append(partBuffer, data[processedLen:processedLen+int64(partSize)]...)
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

func (e *EncryptedChunks) composeOriginalData() ([]byte, error) {
	if len(e.chunks) != len(e.chunksKeys) {
		return nil, fmt.Errorf("unexpected key number having %d requiring %d", len(e.chunksKeys), len(e.chunks))
	}
	// decrypt original data
	outData := make([]byte, 0)
	for idx, edata := range e.chunks {
		key := e.chunksKeys[idx]
		decryptedChunk, err := crypto3n.AesDecrypt(key, edata, crypto3n.CBC)
		if err != nil {
			return nil, err
		}
		outData = append(outData, decryptedChunk...)
	}

	var err error
	// if compressed decompress
	if e.compressed {
		outData, err = ungzipData(outData)
		if err != nil {
			return nil, err
		}
	}

	// check for data size
	if len(outData) != int(e.metadata.Size) {
		return nil, fmt.Errorf("unexpected file size, having %d expecting %d", len(outData), e.metadata.Size)
	}

	return outData, nil
}

func (e *EncryptedChunks) FileFromEncryptedChunks(filepath string) error {
	// get original data
	data, err := e.composeOriginalData()
	if err != nil {
		return err
	}
	// if required untar it
	if e.metadata.IsDir == true {
		err = untar(data, filepath)
		if err != nil {
			return err
		}
	} else {
		// write to file
		err = ioutil.WriteFile(filepath, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewEncryptedChunksFromFile(masterkey []byte, filepath string, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	// get infos from file
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// create chunk struct
	chunk, err := initEncryptedChunks(masterkey, chunkSize, compressed)
	if err != nil {
		return nil, err
	}
	// define metadata
	chunk.metadata.FileName = fileInfo.Name()
	chunk.metadata.IsDir = fileInfo.IsDir()
	chunk.metadata.ModTime = fileInfo.ModTime()

	var data []byte
	if fileInfo.IsDir() == true {
		// create a tmp tar archive and operate on it
		data, err = tarit(filepath)
		if err != nil {
			return nil, err
		}
	} else {
		// read argument passed file
		data, err = ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
	}
	chunk.metadata.Size = int64(len(data))

	// compress data
	if chunk.compressed == true {
		// reassign data
		data = gzipData(data)
	}

	err = chunk.splitDataInChunks(data)
	if err != nil {
		return nil, err
	}

	return chunk, nil
}

func initEncryptedChunks(masterkey []byte, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	if len(masterkey) < minKeyLen {
		return nil, fmt.Errorf("unable to create an encrypted chunk, key is too short: having %d expecting %d", len(masterkey), minKeyLen)
	}
	if chunkSize < minChunkSize {
		return nil, fmt.Errorf("required chunk size is too small: should be major than %d", minChunkSize)
	}
	return &EncryptedChunks{
		masterKey:  masterkey,
		compressed: compressed,
		chunkSize:  chunkSize,
	}, nil
}
