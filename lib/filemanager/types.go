//
// 3nigm4 filemanager package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

package filemanager

// Standard libs
import (
	"crypto/sha512"
	"time"
)

// Internal libs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Metadata metadata related to the original file,
// will be managed locally with encryption keys.
type Metadata struct {
	FileName string               `json:"filename" xml:"filename"`
	Size     int64                `json:"size" xml:"size"`
	ModTime  time.Time            `json:"modtime" xml:"modtime"`
	IsDir    bool                 `json:"isdir" xml:"isdir"`
	CheckSum [sha512.Size384]byte `json:"checksum" xml:"checksum"`
}

// EncryptedChunks encrypted data chunks related
// to their keys, these are the files will be uploaded
// to the cloud storage. All the keys, metadata and
// encryption algorithm details will be saved locally
// only (never passed in plain text anywhere).
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

// ReferenceFile the locally saved output file will
// contain all required info to later on decrypt data
// chunks. If lost there will be no way to recover
// original data.
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

// Permission defines files associated access
// permission.
type Permission struct {
	Permission   ct.Permission
	SharingUsers []string
}

// ContextID is used to pass back to calling
// function a context usable to get progress
// infos while interacting with DataSaver
// interface.
type ContextID string

// ProgressStatus status of a DataSaver managed
// operation, can be used to monitor how the op.
// is going on and how quickly
type ProgressStatus interface {
}

// DataSaver interface of the actual saver for
// encrypted data: this can be a local file system,
// a remote fs or APIs or any other system capable
// of storing data chunks.
type DataSaver interface {
	ProgressStatus(ContextID) (ProgressStatus, error)                                              // Get a requestID argument and return progress infos about;
	SaveChunks(string, [][]byte, []byte, time.Duration, *Permission, *ContextID) ([]string, error) // Saves chunks using a file name, bucket, actual data, a checksum reference and an expire date;
	RetrieveChunks(string, []string, *ContextID) ([][]byte, error)                                 // Retrieve all resources composing a file;
	DeleteChunks(string, []string, *ContextID) error                                               // removes all resources composing a file.
}
