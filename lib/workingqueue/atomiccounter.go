//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

package workingqueue

import (
	"sync"
)

// AtomicCounter concurrent safe counter.
type AtomicCounter struct {
	mutex sync.Mutex
	x     int64
}

// Add increment of quantity x the counter.
func (c *AtomicCounter) Add(x int64) {
	c.mutex.Lock()
	c.x += x
	c.mutex.Unlock()
}

// Value returns the actual counter  value.
func (c *AtomicCounter) Value() int64 {
	c.mutex.Lock()
	val := c.x
	c.mutex.Unlock()
	return val
}
