//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package commons

// Golang std libs
import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RandomBytesForLen creates a random secure data blob
// of length "size". If anything went wrong, or not
// enought entropy is available, the function returns
// an error.
func RandomBytesForLen(size int) ([]byte, error) {
	randData := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, randData); err != nil {
		return nil, err
	}
	return randData, nil
}

// VerifyDestinationPath checks argument file path on different aspects:
// non nullity, base dir existance, non dir and non already existant.
func VerifyDestinationPath(path string) error {
	if path == "" {
		return fmt.Errorf("a output path is required to save the produced QRCode png image")
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		info, err := os.Stat(dir)
		if err != nil ||
			info.IsDir() != true {
			return fmt.Errorf("provided path to output file is invalid, %d do not exist", dir)
		}
	} else {
		return fmt.Errorf("a file named %s already exist, please remove it before use this path", path)
	}

	return nil
}
