//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

package workingqueue

import (
	"sync"
	"testing"
	"time"
)

const (
	workersNumber         = 8
	sleepTimeMilliseconds = 5
	dispatchWorks         = 500
)

func TestDispatcher(t *testing.T) {
	dispatchc := make(chan job)
	errc := make(chan error)

	dispatcher := newDispatcher(workersNumber, errc, dispatchc)
	if dispatcher == nil {
		t.Fatalf("Dispatcher was not created.\n")
	}
	defer dispatcher.stop()

	err := dispatcher.run()
	if err != nil {
		t.Fatalf("Unable to run dispatcher: %s.\n", err.Error())
	}

	// sync vars
	counter := AtomicCounter{}
	var wg sync.WaitGroup

	// manage errors and workerpool
	go func() {
		for {
			select {
			case e := <-dispatcher.errorChan:
				t.Logf("Error while waiting for work: %s.\n", e.Error())
				counter.Add(1)
			}
		}
	}()
	// send messages
	for idx := 0; idx < dispatchWorks*workersNumber; idx++ {
		wg.Add(1)
		go func(i int) {
			arg := Args{
				Data:  []byte("This is some fake data"),
				Index: i,
				Sleep: sleepTimeMilliseconds,
			}
			work := job{
				function: processing,
				args:     arg,
			}
			dispatcher.jobQueue <- work
			wg.Done()
		}(idx)
	}
	wg.Wait()

	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (wg waits only for the chan
	// injection).
	ticker := time.Tick(10 * time.Second)
	timeoutCounter := AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	var processedCount int64
	for {
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Some message missing, having %d expecting %d.\n", processedCount, dispatchWorks*workersNumber)
		}
		if counter.Value() != 0 {
			t.Fatalf("Some error occurred: %d.\n", counter.Value())
		}
		processedCount = 0
		for _, worker := range dispatcher.workers {
			processedCount += worker.countedJobs()
		}
		if processedCount == int64(dispatchWorks*workersNumber) {
			break
		}
		time.Sleep(3 * time.Millisecond)
	}
}
