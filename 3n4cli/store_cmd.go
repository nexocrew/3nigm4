//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"os"
	"sync"
	"time"
)

// Internal dependencies
import (
	fm "github.com/nexocrew/3nigm4/lib/filemanager"
	sc "github.com/nexocrew/3nigm4/lib/storageclient"
)

// Third party libs
import (
	"github.com/sethgrid/multibar"
	"github.com/spf13/cobra"
)

// StoreCmd clinet service that connect to the service API
// to manage sensible data, typically exposed operations
// are upload, download and delete.
var StoreCmd = &cobra.Command{
	Use:       "store",
	Short:     "Store securely data to the cloud",
	Long:      "Store and manage secured data to the colud. All the encryption routines are executed on the client only encrypted chunks are sended to the server.",
	Example:   "3n4cli store",
	ValidArgs: []string{"upload", "download", "delete"},
	RunE:      store,
}

// manageAsyncErrors is a common function used by the various
// store child commands to manage async returned errors. If
// an error is returned exit is invoked.
func manageAsyncErrors(errc <-chan error) {
	for {
		select {
		case err, _ := <-errc:
			log.CriticalLog("Error encountered: %s.\n", err.Error())
			os.Exit(1)
		}
	}
}

// progressBarUpdate function should be invoked concurrently to
// update cli progress bar.
func progressBarUpdate(ctx *fm.ContextID, ds *sc.StorageClient, pf multibar.ProgressFunc, wg *sync.WaitGroup) {
	for {
		if *ctx == "" {
			time.Sleep(time.Millisecond * 15)
			continue
		}
		status, err := ds.ProgressStatus(*ctx)
		if err != nil {
			break
		}
		// x : 100 = progress : total
		pf((100 * status.Done()) / status.TotalUnits())
		if status.Done() == status.TotalUnits() {
			break
		}
		time.Sleep(time.Millisecond * 15)
	}
	wg.Done()
}

// serve command expose a RPC service that exposes all authentication
// related function to the outside.
func store(cmd *cobra.Command, args []string) error {
	return nil
}
