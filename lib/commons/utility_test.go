//
// 3nigm4 commons package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package commons

// Golang std libs
import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	testPathCorrect      = "/tmp/testfile.txt"
	testPathNonExistant  = "/tmp/stranglyinopportunedir/file.txt"
	testPathIsDir        = "/tmp"
	testExistingFilePath = "/tmp/existing.txt"
	fakeData             = "This is a test data string to be saved in fake test file."
)

func TestVerifyDirectoryPath(t *testing.T) {
	err := VerifyDestinationPath("")
	if err == nil {
		t.Fatalf("Empty path should rise an error.\n")
	}
	err = VerifyDestinationPath(testPathNonExistant)
	if err == nil {
		t.Fatalf("The non existant path should rise an error.\n")
	}
	err = VerifyDestinationPath(testPathIsDir)
	if err == nil {
		t.Fatalf("Existant directory path should rise an error.\n")
	}
	err = VerifyDestinationPath(testPathCorrect)
	if err != nil {
		t.Fatalf("Correct path should not rise errors: %s.\n", err.Error())
	}
	err = ioutil.WriteFile(testExistingFilePath, []byte(fakeData), 0600)
	if err != nil {
		t.Fatalf("Unable to write file to %s path.\n", testExistingFilePath)
	}
	defer os.Remove(testExistingFilePath)
	err = VerifyDestinationPath(testExistingFilePath)
	if err == nil {
		t.Fatalf("Existing files should rise an error.\n")
	}
}
