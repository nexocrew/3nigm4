//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package workingqueue

import (
	"fmt"
)

// Dispatcher implement a working queue to optimise
// access to shared messaging queue from the input
// components.
type dispatcher struct {
	workerPool chan chan job // A pool of workers channels that are registered with the dispatcher;
	errorChan  chan error    // The channel to be used to return errors;
	workers    []*worker     // The slice of initialised workers;
	maxWorkers int           // maximum number of workers;
	jobQueue   chan job      // The job queue that serialise jobs;
	quit       chan bool     // stop activity chan.
}

// NewDispatcher creates a new dispatcher object to be
// filled with required workers.
func newDispatcher(maxWorkers int, errorc chan error, jobQueue chan job) *dispatcher {
	return &dispatcher{
		workerPool: make(chan chan job, maxWorkers),
		errorChan:  errorc,
		workers:    make([]*worker, maxWorkers),
		maxWorkers: maxWorkers,
		quit:       make(chan bool),
		jobQueue:   jobQueue,
	}
}

// Run start all workers and put the dispatcher
// in listening mode on the job queue.
func (d *dispatcher) run() error {
	// starting workers
	for idx := 0; idx < d.maxWorkers; idx++ {
		worker, err := newWorker(idx, d.workerPool, d.errorChan)
		if err != nil {
			return err
		}
		worker.start()
		d.workers[idx] = worker
	}
	// start dispatching
	go d.dispatch()
	return nil
}

// dispatch start dispatching jobs to available
// workers.
func (d *dispatcher) dispatch() {
	// define boolean conditions
	jobc_closed := false

	for {
		// check for closed channel
		if jobc_closed == true {
			d.errorChan <- fmt.Errorf("unable to dispatch queue, job channel is closed")
			return
		}
		// select on channels
		select {
		case job, jobc_ok := <-d.jobQueue:
			if !jobc_ok {
				jobc_closed = true
			} else {
				// a job request has been received
				go func() {
					// try to obtain a worker job channel that is available.
					// this will block until a worker is idle
					workerJobQueue := <-d.workerPool
					// dispatch the job
					workerJobQueue <- job
				}()
			}
		case <-d.quit:
			for _, worker := range d.workers {
				worker.stop()
			}
			return
		}
	}
}

// Stop method signals the workers to stop.
func (d *dispatcher) stop() {
	// stop queue consuming
	go func() {
		d.quit <- true
	}()
}