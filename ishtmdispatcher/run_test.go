//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"testing"
)

/*
// Internal dependencies.
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	mockdb "github.com/nexocrew/3nigm4/lib/ishtm/mocks"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
)

func TestMain(m *testing.M) {
	// start up logging facility
	log = logger.NewLogFacility("ishtmdispatcher", true, true)

	arguments = args{
		verbose:     true,
		colored:     true,
		dbAddresses: fmt.Sprintf("%s:%d", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:  itm.S().DbUserName(),
		dbPassword:  itm.S().DbPassword(),
		dbAuth:      itm.S().DbAuth(),
		address:     mockServiceAddress,
		port:        mockServicePort,
	}
	databaseStartup = mockDbStartup

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
	// startup service
	go func(ec chan error) {
		err := serve(ServeCmd, nil)
		if err != nil {
			ec <- err
			return
		}
	}(errorChan)
	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (especially to verify that
	// no error is returned by concurrent server
	// startup). 3 seconds is an arbitrary, experimentally
	// defined, time on some slower systems it can be not
	// enought.
	ticker := time.Tick(3 * time.Second)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	// infinite loop:
	for {
		if timeoutCounter.Value() != 0 {
			break
		}
		if errorCounter.Value() != 0 {
			log.ErrorLog("Error returned: %s.\n", lastError)
			os.Exit(1)
		}
		time.Sleep(50 * time.Millisecond)
	}

	os.Exit(m.Run())
}
*/
func TestProcessingFlow(t *testing.T) {

}
