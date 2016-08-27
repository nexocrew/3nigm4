//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package workingqueue

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// messages target
const worksNumber = 500

type Args struct {
	Data   []byte
	Index  int
	Sleep  time.Duration
	Result chan string
}

func processing(args interface{}) error {
	var argument Args
	var ok bool
	if argument, ok = args.(Args); !ok {
		return fmt.Errorf("unknown argument type, having %s expecting type Args", reflect.TypeOf(args))
	}
	resulting := fmt.Sprintf("Case %d processed: \"%s\"", argument.Index, string(argument.Data))
	if argument.Result != nil {
		argument.Result <- resulting
	}
	time.Sleep(argument.Sleep * time.Millisecond)
	return nil
}

func TestWorker(t *testing.T) {
	workerPool := make(chan chan job)
	errc := make(chan error)
	resultc := make(chan string)

	worker, err := newWorker(0, workerPool, errc)
	if err != nil {
		t.Fatalf("Unable to create worker: %s.\n", err.Error())
	}
	worker.start()
	defer worker.stop()
	// sync vars
	counter := AtomicCounter{}
	processedc := AtomicCounter{}
	resultingc := AtomicCounter{}
	var wg sync.WaitGroup

	var resultingStrings []string

	// manage errors and workerpool
	go func() {
		for {
			select {
			case <-errc:
				counter.Add(1)
			case <-worker.workerPool:
				processedc.Add(1)
			case str, _ := <-resultc:
				resultingc.Add(1)
				resultingStrings = append(resultingStrings, str)
			}
		}
	}()
	// send messages
	for idx := 0; idx < worksNumber; idx++ {
		wg.Add(1)
		go func(i int) {
			arg := Args{
				Data:   []byte("This is some fake data"),
				Index:  i,
				Result: resultc,
			}
			work := job{
				function: processing,
				args:     arg,
			}
			worker.jobChannel <- work
			wg.Done()
		}(idx)
	}
	wg.Wait()

	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (wg waits only for the chan
	// injection).
	ticker := time.Tick(1 * time.Second)
	timeoutCounter := AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	for {
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Unexpected number of processed messages: having %d expecting %d.\n", processedc.Value(), worksNumber+1)
		}
		if counter.Value() != 0 {
			t.Fatalf("Some error occurred: %d.\n", counter.Value())
		}
		// check with +1 cause workerpool is passed before
		// blocking in the switch block (so it we'll be done
		// even after reciving the last message).
		if processedc.Value() == worksNumber+1 &&
			resultingc.Value() == worksNumber {
			break
		}
		time.Sleep(3 * time.Millisecond)
	}
	if len(resultingStrings) != worksNumber {
		t.Fatalf("Unexpected number of results having %d expecting %d.\n", len(resultingStrings), worksNumber)
	}
	for i, value := range resultingStrings {
		var selfc uint
		for r, reference := range resultingStrings {
			if value == reference {
				if selfc > 0 {
					t.Fatalf("Founded two equal strings: %s (%d) == %s (%d)", value, i, reference, r)
				} else {
					selfc++
				}
			}
		}
	}
}

func TestWorkerWithError(t *testing.T) {
	workerPool := make(chan chan job)
	errc := make(chan error)
	resultc := make(chan string)

	worker, err := newWorker(0, workerPool, errc)
	if err != nil {
		t.Fatalf("Unable to create worker: %s.\n", err.Error())
	}
	worker.start()
	defer worker.stop()
	// sync vars
	counter := AtomicCounter{}
	processedc := AtomicCounter{}
	resultingc := AtomicCounter{}
	var wg sync.WaitGroup

	var resultingStrings []string

	// manage errors and workerpool
	go func() {
		for {
			select {
			case <-errc:
				counter.Add(1)
			case <-worker.workerPool:
				processedc.Add(1)
			case str, _ := <-resultc:
				resultingc.Add(1)
				resultingStrings = append(resultingStrings, str)
			}
		}
	}()
	// send messages
	for idx := 0; idx < worksNumber; idx++ {
		wg.Add(1)
		go func(i int) {
			var generic interface{}
			if i == 10 ||
				i == 3 ||
				i == 15 {
				generic = "This is a wrong message type"
			} else {
				generic = Args{
					Data:   []byte("This is some fake data"),
					Index:  i,
					Result: resultc,
				}
			}
			work := job{
				function: processing,
				args:     generic,
			}
			worker.jobChannel <- work
			wg.Done()
		}(idx)
	}
	wg.Wait()

	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (wg waits only for the chan
	// injection).
	ticker := time.Tick(1 * time.Second)
	timeoutCounter := AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	for {
		if timeoutCounter.Value() != 0 {
			t.Fatalf("Unexpected number of processed messages: having %d expecting %d.\n", processedc.Value(), worksNumber+1)
		}
		// check with +1 cause workerpool is passed before
		// blocking in the switch block (so it we'll be done
		// even after reciving the last message).
		if processedc.Value() == worksNumber+1 {
			break
		}
		time.Sleep(3 * time.Millisecond)
	}
	if counter.Value() != 3 {
		t.Fatalf("Expected %d errors but found %d.\n", 3, counter.Value())
	}
	if resultingc.Value() != worksNumber-3 {
		t.Fatalf("Expected returning %d correctly processed jobs but founded %d.\n", worksNumber-3, resultingc.Value())
	}
}
