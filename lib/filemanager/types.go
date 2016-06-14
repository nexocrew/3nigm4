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
	compressed bool
	chunkSize  uint64
	chunks     [][]byte
	chunksKeys [][]byte
	metadata   Metadata
	// optional master key
	masterKey        []byte
	derivationRounds int
	salt             []byte
}

type ChunkFile struct {
	Chunk   []byte    `json:"chunk" xml:"chunk"`
	ModTime time.Time `json:"modtime" xml:"modtime"`
}

type ReferenceFile struct {
	// metadata
	FileName string               `json:"filename" xml:"filename"`
	Size     int64                `json:"size" xml:"size"`
	ModTime  time.Time            `json:"modtime" xml:"modtime"`
	IsDir    bool                 `json:"isdir" xml:"isdir"`
	CheckSum [sha512.Size384]byte `json:"checksum" xml:"checksum"`
	// encryption
	DerivationRounds int      `json:"rounds" xml:"rounds"`
	Salt             []byte   `json:"salt" xml:"salt"`
	ChunksKeys       [][]byte `json:"chunkskeys" xml:"chunkskeys"`
	// chunks settings
	ChunksPaths []string `json:"chunkspaths" xml:"chunkspaths"`
	Compressed  bool     `json:"compressed" xml:"compressed"`
	ChunkSize   uint64   `json:"chunksize" xml:"chunksize"`
}

type DataSaver interface {
	SaveChunks(string, [][]byte) ([]string, error)
	RetrieveChunks([]string) ([][]byte, error)
}
