//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package commons

// Golang std libs
import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var filename string

func TestMain(m *testing.M) {
	fd, err := ioutil.TempFile("", "testfiles")
	if err != nil {
		fmt.Printf("Error creating tmp file: %s.\n", err.Error())
		os.Exit(1)
	}

	fakeTestData, err := RandomBytesForLen(3145728)
	if err != nil {
		fmt.Printf("Error generating random data: %s.\n", err.Error())
		os.Exit(1)
	}
	if _, err := fd.Write(fakeTestData); err != nil {
		fmt.Printf("Error writing data to tmp file: %s.\n", err.Error())
		os.Exit(1)
	}
	if err := fd.Close(); err != nil {
		fmt.Printf("Error closing tmp file: %s.\n", err.Error())
		os.Exit(1)
	}
	filename = fd.Name()

	os.Exit(m.Run())
}

func cleanTestFile() error {
	os.Remove(filename)
	return nil
}

func TestPass1(t *testing.T) {
	if filename == "" {
		t.Fatalf("Unexpected empty file name.\n")
	}
	fd, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}

	err = passImplementation(fd, Pass1)
	if err != nil {
		t.Fatalf("Unable to implement pass: %s.\n", err.Error())
	}
	fd.Close()

	fd, err = os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		t.Fatalf("Unable to read stats: %s.\n", err.Error())
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = fd.Read(buffer)
	if err != nil {
		t.Fatalf("Unable to read file: %s.\n", err.Error())
	}
	for idx := 0; idx < len(buffer); idx++ {
		if buffer[idx] != byte(0x0) {
			t.Fatalf("Unexpected byte at index %d, having %v expecting %v.\n", idx, buffer[idx], byte(0x0))
		}
	}
}

func TestPass2(t *testing.T) {
	if filename == "" {
		t.Fatalf("Unexpected empty file name.\n")
	}
	fd, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}

	err = passImplementation(fd, Pass2)
	if err != nil {
		t.Fatalf("Unable to implement pass: %s.\n", err.Error())
	}
	fd.Close()

	fd, err = os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		t.Fatalf("Unable to read stats: %s.\n", err.Error())
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = fd.Read(buffer)
	if err != nil {
		t.Fatalf("Unable to read file: %s.\n", err.Error())
	}
	for idx := 0; idx < len(buffer); idx++ {
		if buffer[idx] != byte(0x1) {
			t.Fatalf("Unexpected byte at index %d, having %v expecting %v.\n", idx, buffer[idx], byte(0x0))
		}
	}
}

func TestPass3(t *testing.T) {
	if filename == "" {
		t.Fatalf("Unexpected empty file name.\n")
	}
	fd, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}

	err = passImplementation(fd, Pass3)
	if err != nil {
		t.Fatalf("Unable to implement pass: %s.\n", err.Error())
	}
	fd.Close()

	fd, err = os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		t.Fatalf("Unable to open file: %s.\n", err.Error())
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		t.Fatalf("Unable to read stats: %s.\n", err.Error())
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = fd.Read(buffer)
	if err != nil {
		t.Fatalf("Unable to read file: %s.\n", err.Error())
	}
	var nonEntropyCount int
	for idx := 0; idx < len(buffer); idx++ {
		if nonEntropyCount >= 4 {
			t.Fatalf("Too poor entropy, seems the buffer has not been filled with zeroes.\n")
		}
		if buffer[idx] == byte(0x1) {
			nonEntropyCount++
		} else {
			nonEntropyCount = 0
		}
	}
}
