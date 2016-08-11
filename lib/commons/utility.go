//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package commons

// Golang std libs
import (
	"crypto/rand"
	"io"
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
