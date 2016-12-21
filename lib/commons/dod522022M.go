//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 19/12/2016
//

package commons

// Golang std libs
import (
	"fmt"
	"math"
	"os"
)

type Pass uint

const (
	Pass1 Pass = iota
	Pass2 Pass = iota
	Pass3 Pass = iota
)

func bufferWithByte(value byte, size int) []byte {
	buffer := make([]byte, size)
	for idx := 0; idx < size; idx++ {
		buffer[idx] = value
	}
	return buffer
}

func passImplementation(fd *os.File, pass Pass) error {
	fileInfo, err := fd.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	lastPosition := 0
	for idx := uint64(0); idx < totalPartsNum; idx++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(idx*fileChunk))))
		var overridingBytes []byte
		switch pass {
		case Pass1:
			overridingBytes = bufferWithByte(byte(0x0), partSize)
		case Pass2:
			overridingBytes = bufferWithByte(byte(0x1), partSize)
		case Pass3:
			randomData, err := RandomBytesForLen(partSize)
			if err != nil {
				return err
			}
			overridingBytes = randomData
		}
		n, err := fd.WriteAt(overridingBytes, int64(lastPosition))
		if err != nil {
			return err
		}
		if n != partSize {
			return fmt.Errorf("unexpected part size, having %d expecting %d", partSize, n)
		}
		// update last written position
		lastPosition = lastPosition + partSize
	}
	return nil
}

// SecureFileWipe securely wipe a file from the file system, uses
// DOD5220_22M as algorithm to completely remove the file.
func SecureFileWipe(fd *os.File) error {
	if fd == nil {
		return fmt.Errorf("argument file is nil")
	}
	for pass := 0; pass < 3; pass++ {
		passImplementation(fd, Pass(pass))
	}
	// finally remove/delete our file
	err := os.Remove(fd.Name())
	if err != nil {
		return err
	}
	return nil

}
