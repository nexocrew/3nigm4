//
// 3nigm4 workingqueue package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

// Package workingqueue implement a concurrent working queue
// able to process any passed payload (having a standard
// function signature) managing the maximum number of active
// workers. This is intended to be used on high volumes of
// processing to manage efficiently workloads without doing
// auto Ddos creating always new goroutines.
package workingqueue

// WorkingQueue base struct used to
// represent the working queue.
type WorkingQueue struct {
	jobQueue   chan job    // A buffered channel that we can send work requests on;
	dispatcher *dispatcher // The work dispatcher.
}

// NewWorkingQueue creates a new wq initialising
// all internal properties.
func NewWorkingQueue(workerSize int, queueSize int, errorc chan error) *WorkingQueue {
	// init and alloc the queue
	jobQueue := make(chan job, queueSize)

	// create the producer
	return &WorkingQueue{
		jobQueue:   jobQueue,
		dispatcher: newDispatcher(workerSize, errorc, jobQueue),
	}
}

// Run start dispatching queue.
func (w *WorkingQueue) Run() error {
	return w.dispatcher.run()
}

// Close stop all operations.
func (w *WorkingQueue) Close() {
	w.dispatcher.stop()
}

// MessageCounter returns the number of produced
// messages.
func (w *WorkingQueue) MessageCounter() int64 {
	var counter int64
	for _, wrk := range w.dispatcher.workers {
		counter += wrk.countedJobs()
	}
	return counter
}

// SendJob enqueue a new job in the
// producer queue.
func (w *WorkingQueue) SendJob(payload func(interface{}) error, arguments interface{}) {
	job := job{
		function: payload,
		args:     arguments,
	}
	go func() {
		w.jobQueue <- job
	}()
}
