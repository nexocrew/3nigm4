//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import ()

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
	SetMasterEncryptionKey([]byte) error
	GetFileMetadata() *Metadata
}

type encryptedChunk struct {
	masterKey  []byte
	compressed bool
	chunkSize  int64
	chunks     [][]byte
	chunksKeys [][]byte
	metadata   Metadata
}
