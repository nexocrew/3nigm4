//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

package workingqueue

// Std golang libs
import (
	"fmt"
)

// job to be done
type job struct {
	function func(interface{}) error
	args     interface{}
}

// worker represent the worker that
// executes the job.
type worker struct {
	workerPool   chan chan job
	jobChannel   chan job
	errorChannel chan error
	quit         chan bool
	counter      AtomicCounter
	id           int
}

// newWorker creates a new worker and
// init internal properties.
func newWorker(id int, workerPool chan chan job, errorChan chan error) (*worker, error) {
	return &worker{
		workerPool:   workerPool,
		jobChannel:   make(chan job),
		errorChannel: errorChan,
		quit:         make(chan bool),
		id:           id,
	}, nil
}

// countedJobs returns the number of counted
// processed jobs.
func (w *worker) countedJobs() int64 {
	return w.counter.Value()
}

// start method starts the run loop for the worker,
// listening for a quit channel in case we need to
// stop it.
func (w *worker) start() {
	go func() {
		jobcClosed := false
		for {
			// check for closed channel
			if jobcClosed == true {
				w.errorChannel <- fmt.Errorf("unable to dispatch queue, job channel is closed")
				return
			}
			// register the current worker into the
			// working queue
			w.workerPool <- w.jobChannel
			// manage channels
			select {
			case job, jobcOk := <-w.jobChannel:
				if !jobcOk {
					jobcClosed = true
				} else {
					// worker recived a job
					err := job.function(job.args)
					if err != nil {
						w.errorChannel <- fmt.Errorf("unable to process job cause %s", err.Error())
					} else {
						go func() {
							w.counter.Add(1)
						}()
					}
				}
			case <-w.quit:
				// releasing and stopping the worker
				return
			}
		}
	}()
}

// stop method signals the worker to stop
// listening for work requests.
func (w *worker) stop() {
	go func() {
		w.quit <- true
	}()
}
