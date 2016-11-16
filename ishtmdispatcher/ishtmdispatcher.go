//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"os"
)

// Internal pkgs
import (
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/logo"
	ver "github.com/nexocrew/3nigm4/lib/version"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

// Third party pkgs
import (
	"github.com/spf13/cobra"
)

// Logger global instance
var log *logger.LogFacility

// Global working queue
var workingQueue *wq.WorkingQueue
var errc chan error

const (
	workersize = 32
	queuesize  = 300
)

// Cobra parsed arguments
var arguments args

// RootCmd is the base command used by cobra in the ishtmservice
// exec.
var RootCmd = &cobra.Command{
	Use:   "ishtmdispatcher",
	Short: "If Something Happens To Me dispatching backbone",
	Long:  "Server that manage dispatch of \"will\" records.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo()
		// Execution implementation
		return fmt.Errorf("undefined command, select a valid one")
	},
}

func init() {
	// global flags
	RootCmd.PersistentFlags().BoolVarP(&arguments.verbose, "verbose", "v", false, "activate logging verbosity")
	RootCmd.PersistentFlags().BoolVarP(&arguments.colored, "colored", "C", true, "activate colored logs")
}

// Execute parsing and execute selected
// command.
func Execute() error {
	// execute actual command
	_, err := RootCmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("Ishtm dispatching backbone", ver.V().VersionString(), nil))
}

func main() {
	// start up logging facility
	log = logger.NewLogFacility("ishtmdispatcher", true, true)

	// create working queue
	errc = make(chan error, workersize)
	go func() {
		for {
			select {
			case err := <-errc:
				log.ErrorLog("Async error: %s.\n", err.Error())
			}
		}
	}()
	workingQueue = wq.NewWorkingQueue(workersize, queuesize, errc)
	// start working queue
	if err := workingQueue.Run(); err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}
	defer workingQueue.Close()

	err := Execute()
	if err != nil {
		log.CriticalLog("%s.\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
