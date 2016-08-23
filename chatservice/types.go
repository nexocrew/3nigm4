package main

// Arguments management struct.
type args struct {
	// server basic args
	verbose bool
	port    int
	// https certificates
	sslcrt string
	sslpvk string
	// workers and queues
	workers int
	queue   int
}
