//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package filemanager

// Standard libs
import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func tarit(source string) ([]byte, error) {
	// create a buffer
	buf := new(bytes.Buffer)
	// create a tar
	tarball := tar.NewWriter(buf)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})

	return buf.Bytes(), nil
}

func untar(tarball []byte, target string) error {
	buf := bytes.NewReader(tarball)
	tarReader := tar.NewReader(buf)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func ungzipData(compressed []byte) ([]byte, error) {
	reader := bytes.NewReader(compressed)
	r, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	io.Copy(&out, r)
	return out.Bytes(), nil
}

func gzipData(data []byte) []byte {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	defer w.Close()

	// write in buffer
	w.Write(data)
	w.Close()

	return buf.Bytes()
}

// Returns a random number in the required
// min-max range using a pre-seeded pseudo
// random generator. Not to be used in security
// related functions!
// Produced value is x < n (not equal).
func randomInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}
