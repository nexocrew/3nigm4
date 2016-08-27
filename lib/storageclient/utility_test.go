// 3nigm4 storageclient package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 15/08/2016

package storageclient

// Std golang dependencies.
import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Internal dependencies
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

type serviceStorage struct {
	mtx     sync.Mutex
	storage map[string][]byte
	jobs    map[string]string
}

type safeDelayCounters struct {
	mtx      sync.Mutex
	counters map[string]*wq.AtomicCounter
}

func newSafeDelayCounters() *safeDelayCounters {
	return &safeDelayCounters{
		counters: make(map[string]*wq.AtomicCounter),
	}
}

func (s *safeDelayCounters) initCounter(key string) {
	s.mtx.Lock()
	_, ok := s.counters[key]
	if !ok {
		s.counters[key] = &wq.AtomicCounter{}
	}
	s.mtx.Unlock()
}

func (s *safeDelayCounters) value(key string) int64 {
	s.mtx.Lock()
	value := s.counters[key].Value()
	s.mtx.Unlock()
	return value
}

func (s *safeDelayCounters) add(key string, x int64) {
	s.mtx.Lock()
	s.counters[key].Add(x)
	s.mtx.Unlock()
}

func randomID() (string, error) {
	now := time.Now()
	timeString := fmt.Sprintf("%d.%d", now.Unix(), now.Nanosecond())
	var id []byte
	id = append(id, []byte(timeString)...)
	rand, err := ct.RandomBytesForLen(32)
	if err != nil {
		return "", err
	}
	id = append(id, rand...)
	checksum := sha512.Sum384(id)
	return hex.EncodeToString(checksum[:]), nil
}

func extractAddressAndPort(addr string, t *testing.T) (string, int) {
	components := strings.Split(addr, ":")
	if len(components) != 3 {
		t.Fatalf("Unexpected mock server address, having %d elements instead of 3.\n", len(components))
	}
	address := fmt.Sprintf("%s:%s", components[0], components[1])
	port, err := strconv.Atoi(components[2])
	if err != nil {
		t.Fatalf("Unable to parse int: %s.\n", err.Error())
	}
	return address, port
}
