//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import ()

type FileManager interface {
	WriteFile([]byte) ([]byte, error)
	ReadFile([]byte) ([]byte, error)
}
