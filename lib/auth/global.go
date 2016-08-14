//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// Global variable is used to globally access a unique
// db client (that will be copyied by all functions).
//

package auth

import (
	"sync"
)

// Global vars protecting mutex.
var mtx sync.Mutex

// Runtime allocated global base database instance.
var dbclient Database

// SetGlobalDbClient must be called to set the global db client,
// that implements the Database interface, to be used by RPC
// exposed functions. This function must be always invoked before
// proceeding registering other fucntions.
func SetGlobalDbClient(database Database) {
	mtx.Lock()
	dbclient = database
	mtx.Unlock()
}

// CloseGlobalDbClient closes the global assigned client.
func CloseGlobalDbClient() {
	mtx.Lock()
	dbclient.Close()
	mtx.Unlock()
}
