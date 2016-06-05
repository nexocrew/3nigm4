//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import (
	"crypto/sha512"
	"time"
)

type Metadata struct {
	FileName string               `json:"filename" xml:"filename"`
	Size     int64                `json:"size" xml:"size"`
	ModTime  time.Time            `json:"modtime" xml:"modtime"`
	IsDir    bool                 `json:"isdir" xml:"isdir"`
	CheckSum [sha512.Size384]byte `json:"checksum" xml:"checksum"`
}

type EncryptedChunks struct {
	masterKey  []byte
	compressed bool
	chunkSize  uint64
	chunks     [][]byte
	chunksKeys [][]byte
	metadata   Metadata
}

type ReferenceFile struct {
	Metadata
	ChunksPaths []string `json:"chunkspaths" xml:"chunkspaths"`
	ChunksKeys  []byte   `json:"chunkskeys" xml:"chunkskeys"`
}
