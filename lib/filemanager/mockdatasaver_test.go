//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

import (
	"fmt"
	"os"
	"testing"
)

type localDataSaver struct {
	rootPath string
}

func NewLocalDataSaver(root string) (*localDataSaver, error) {
	if fi, err := os.Stat(root); os.IsNotExist(err) ||
		fi.IsDir() != true {
		return nil, fmt.Error("invalid root path")
	}
	return &localDataSaver{
		rootPath: root,
	}
}

func (l *localDataSaver) SaveChunks(filename string, chunks [][]byte) ([]string, error) {

}

func (l *localDataSaver) RetrieveChunks(files []string) ([][]byte, error) {

}
