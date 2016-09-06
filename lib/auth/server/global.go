//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// Global variable is used to globally access a unique
// db client (that will be copyied by all functions).
//

package authserver

import (
	"sync"
)

import (
	db "github.com/nexocrew/3nigm4/lib/database/client"
)

// Global vars protecting mutex.
var mtx sync.Mutex

// Runtime allocated global base database instance.
var dbclient db.Database

// SetGlobalDbClient must be called to set the global db client,
// that implements the Database interface, to be used by RPC
// exposed functions. This function must be always invoked before
// proceeding registering other fucntions.
func SetGlobalDbClient(database db.Database) {
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
