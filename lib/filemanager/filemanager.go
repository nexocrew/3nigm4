//
// 3nigm4 filemanager package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

// Package filemanager manage the split and encryption of a
// file in chunkes that will be uploadable to a cloud storage.
// This package is part of the security strategy of 3nigm4: dividing
// the file in chunks and assigning a unique resource id will
// produce unrelated anonymous chunks that cannot be related with
// the original file metadata (lenght, hash and so on...).
package filemanager

// Standard libs
import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	mathrand "math/rand"
	"os"
	"time"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

const (
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

func (e *EncryptedChunks) defineKeyAndSaltForIdx(idx uint64) ([]byte, []byte, error) {
	var key, salt []byte
	var err error
	if e.masterKey != nil &&
		len(e.masterKey) == chunkKeySize &&
		e.salt != nil {
		// xor random key with derived
		key, err = crypto3n.XorKeys([][]byte{e.chunksKeys[idx], e.masterKey}, chunkKeySize)
		if err != nil {
			return nil, nil, err
		}
		salt = e.salt
	} else {
		// assign random key
		key = e.chunksKeys[idx]
		// generate random salt of len 8 bytes
		salt = make([]byte, 8)
		_, err := rand.Read(salt)
		if err != nil {
			return nil, nil, err
		}
	}
	return key, salt, nil
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

	var processedLen uint64
	for idx := uint64(0); idx < totalPartsCount; idx++ {
		remainingLen := uint64(len(data)) - processedLen
		partSize := uint64(math.Min(float64(e.chunkSize), float64(remainingLen)))
		var partBuffer []byte
		// copy data blob
		partBuffer = append(partBuffer, data[processedLen:processedLen+partSize]...)
		processedLen += partSize

		key, salt, err := e.defineKeyAndSaltForIdx(idx)
		if err != nil {
			return err
		}
		// encrypt using associated key
		encryptedChunk, err := crypto3n.AesEncrypt(key, salt, partBuffer, crypto3n.CBC)
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
	var outData []byte
	for idx, edata := range e.chunks {
		var key []byte
		var err error
		// xor with master key if any
		key, _, err = e.defineKeyAndSaltForIdx(uint64(idx))
		if err != nil {
			return nil, err
		}
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

func deriveAesMasterKey(rawKey []byte, rounds int, salt []byte) ([]byte, []byte, error) {
	var s []byte
	if salt != nil &&
		len(salt) == 8 {
		s = salt
	} else {
		// randomly generate salt
		s = make([]byte, 8)
		_, err := rand.Read(s)
		if err != nil {
			return nil, nil, err
		}
	}

	key := crypto3n.DeriveKeyWithPbkdf2(rawKey, s, rounds)
	return key, s, nil
}

func initEncryptedChunks(rawkey []byte, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	if chunkSize < minChunkSize {
		return nil, fmt.Errorf("required chunk size is too small: should be major than %d", minChunkSize)
	}

	ec := &EncryptedChunks{
		compressed: compressed,
		chunkSize:  chunkSize,
	}

	// if masterkey available
	if rawkey != nil &&
		len(rawkey) > 0 {
		// define rounds
		r := mathrand.New(mathrand.NewSource(time.Now().Unix()))
		ec.derivationRounds = randomInRange(r, 10000, 13000)
		key, salt, err := deriveAesMasterKey(rawkey, ec.derivationRounds, nil)
		if err != nil {
			return nil, err
		}
		ec.masterKey = key
		ec.salt = salt
	}
	return ec, nil
}

// NewEncryptedChunks creates a new encrypted chunks
// structure from a given file, a chunk size and a compression
// flag. If a rawkey is specified will be used to make AES
// encryption stronger (this key will not be passed using a
// reference file).
// This function returns the initialised struct or an error if
// sometring went wrong.
func NewEncryptedChunks(rawKey []byte, filepath string, chunkSize uint64, compressed bool) (*EncryptedChunks, error) {
	// get infos from file
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// create chunk struct
	chunk, err := initEncryptedChunks(rawKey, chunkSize, compressed)
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

	// calculate checksum
	chunk.metadata.CheckSum = sha512.Sum384(data)

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

// SaveChunks saves encrypted data chunks to
// a structure implementing the DataSaver interface.
func (e *EncryptedChunks) SaveChunks(ds DataSaver, bucket string, expires *time.Time) (*ReferenceFile, error) {
	filesPaths, err := ds.SaveChunks(e.metadata.FileName, bucket, e.chunks, e.metadata.CheckSum[:], expires)
	if err != nil {
		return nil, err
	}

	rf := &ReferenceFile{
		// metadata
		FileName: e.metadata.FileName,
		Size:     e.metadata.Size,
		CheckSum: e.metadata.CheckSum,
		ModTime:  e.metadata.ModTime,
		IsDir:    e.metadata.IsDir,
		// encryption vars
		DerivationRounds: e.derivationRounds,
		Salt:             e.salt,
		ChunksKeys:       e.chunksKeys,
		// file paths
		ChunksPaths: filesPaths,
		ChunkSize:   e.chunkSize,
		Compressed:  e.compressed,
	}
	return rf, nil
}

// LoadChunks loads chunks from a struct implementing
// the DataSaver interface, given a reference file in
// input. It returns a complete encrypted chunks structure
// from which decrypt the original file.
func LoadChunks(ds DataSaver, reference *ReferenceFile, rawKey []byte) (*EncryptedChunks, error) {
	chunks, err := ds.RetrieveChunks(reference.ChunksPaths)
	if err != nil {
		return nil, err
	}

	var key []byte
	if rawKey != nil &&
		len(rawKey) != 0 {
		key, _, err = deriveAesMasterKey(rawKey, reference.DerivationRounds, reference.Salt)
	}

	if err != nil {
		return nil, err
	}
	ec := &EncryptedChunks{
		metadata: Metadata{
			FileName: reference.FileName,
			Size:     reference.Size,
			ModTime:  reference.ModTime,
			IsDir:    reference.IsDir,
			CheckSum: reference.CheckSum,
		},
		chunkSize:        reference.ChunkSize,
		compressed:       reference.Compressed,
		chunks:           chunks,
		chunksKeys:       reference.ChunksKeys,
		derivationRounds: reference.DerivationRounds,
		salt:             reference.Salt,
		masterKey:        key,
	}

	return ec, nil
}

// GetFile returns the recomposed file merging all
// data chunks and verifying consistency. It saves
// the final result to the path specified as argument
// or returns an error is something went wrong.
func (e *EncryptedChunks) GetFile(filepath string) error {
	// get original data
	data, err := e.composeOriginalData()
	if err != nil {
		return err
	}

	// checksum verification
	actualCs := sha512.Sum384(data)
	if bytes.Compare(actualCs[:], e.metadata.CheckSum[:]) != 0 {
		return fmt.Errorf("checksum not verified, hashed value from actual data do not match reference, file malformed")
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

// ChunkFileId calculate the file name for a specific chunk and
// returns an hexed string that should be used to store it in a
// data saver implementation. Checksum data can be any hased data
// usable to differentiate commonly named files (being derivable
// form metadata).
func ChunkFileId(filename string, chunkNumber int, checksum []byte) (string, error) {
	completeFileName := fmt.Sprintf("%s-chunk%d", filename, chunkNumber)
	var id []byte
	id = append(id, []byte(completeFileName)...)
	id = append(id, checksum...)
	hashedId := sha1.Sum(id)
	return hex.EncodeToString(hashedId[:]), nil
}
