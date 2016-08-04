//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

/*
// Std Golang libs
import (
	"fmt"
	"testing"
)

// Internal dependencies.
import (
	"github.com/nexocrew/3nigm4/lib/itm"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

func TestRPCServe(t *testing.T) {
	arguments = args{
		verbose:     true,
		colored:     true,
		dbAddresses: fmt.Sprintf("%s:%s", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:  itm.S().DbUserName(),
		dbPassword:  itm.S().DbPassword(),
		dbAuth:      itm.S().DbAuth(),
	}

	var errorCounter wq.AtomicCounter
	errorChan := make(chan error, 0)
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-errorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()

	go func(ec chan error) {
		err := serve(ServeCmd, nil)
		if err != nil {
			ec <- err
		}
	}(errorChan)

	for {
		if errorCounter.Value() != 0 {
			t.Fatalf("Error returned: %s.\n", lastError)
		}
	}

}
*/
