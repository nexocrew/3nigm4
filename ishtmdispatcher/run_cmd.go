//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	ishtmdb "github.com/nexocrew/3nigm4/lib/ishtm/db"
	wl "github.com/nexocrew/3nigm4/lib/ishtm/will"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

// Third party pkgs
import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(RunCmd)
}

// ServeCmd start the http/https server listening
// on exec args, used to register to cobra lib root
// command.
var RunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run dispatcher routine",
	Long:    "Launch dispatching routine to loop on the db and send expired \"will\"",
	Example: "ishtmdispatcher run -d 127.0.0.1:27017 -u dbuser -w dbpwd --smtpaddress 192.168.0.1 --smtpport 443 --smtpuser username --smtppwd pwd -v",
}

func init() {
	// database references
	RunCmd.PersistentFlags().StringVarP(&arguments.dbAddresses, "dbaddrs", "d", "127.0.0.1:27017", "the database cluster addresses")
	RunCmd.PersistentFlags().StringVarP(&arguments.dbUsername, "dbuser", "u", "", "the database user name")
	RunCmd.PersistentFlags().StringVarP(&arguments.dbPassword, "dbpwd", "w", "", "the database password")
	RunCmd.PersistentFlags().StringVarP(&arguments.dbAuth, "dbauth", "", "admin", "the database auth db")
	// smtp coordinates
	RunCmd.PersistentFlags().StringVarP(&arguments.senderAddress, "smtpaddress", "", "", "the smtp service address")
	RunCmd.PersistentFlags().IntVarP(&arguments.senderPort, "smtpport", "", 443, "the smtp service port")
	RunCmd.PersistentFlags().StringVarP(&arguments.senderAuthUser, "smtpuser", "", "", "the smtp service user name")
	RunCmd.PersistentFlags().StringVarP(&arguments.senderAuthPassword, "smtppwd", "", "", "the smtp service password")
	// files parameters
	RunCmd.RunE = run
}

// Global database referring variable to be copied and released by
// each goroutine.
var db ct.Database

// Global working queue
var workingQueue *wq.WorkingQueue
var errc chan error

// This var is used to permitt to switch to mock db implementation
// in unit-tests, do not mess with it for other reasons.
// The default, production targeting, implementation uses Mongodb
// as backend database system.
var databaseStartup func(*args) (ct.Database, error) = mgoStartup

// This var is used to permitt to switch between different delivery
// systems, should be settled before proceeding invoking it.
var deliverWill func(*wl.Will) error = nil

// mgoStartup implement startup logic for a mongodb based database
// connection.
func mgoStartup(a *args) (ct.Database, error) {
	// startup db
	mgodb, err := ishtmdb.MgoSession(&ct.DbArgs{
		Addresses: strings.Split(a.dbAddresses, ","),
		User:      a.dbUsername,
		Password:  a.dbPassword,
		AuthDb:    a.dbAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start db connection cause %s", err.Error())
	}

	log.MessageLog("Mongodb %s successfully connected.\n", a.dbAddresses)

	// ensure indexes
	err = mgodb.EnsureMongodbIndexes()
	if err != nil {
		log.WarningLog("Ensuring indexes in Mongodb returned error %s.\n", err.Error())
	}
	return mgodb, nil
}

const (
	workersize   = 16
	queuesize    = 300
	sleepingTime = 3 * time.Minute
)

// startupChans start error chan to manage async
// errors.
func startupChans() {
	errc = make(chan error, workersize)
	go func() {
		for {
			select {
			case err := <-errc:
				log.ErrorLog("Async error: %s.\n", err)
			}
		}
	}()
}

// procArgs processing func used arguments.
type procArgs struct {
	database     ct.Database
	requestTime  time.Time
	errorChan    chan error
	deliveryFunc func(*wl.Will) error
}

// processing execute the actual async processing flow
// it is passed to the working queue and provided with all
// needed arguments.
func processing(genericArgs interface{}) error {
	args, ok := genericArgs.(*procArgs)
	if !ok {
		return fmt.Errorf("unexpected arguments, having %s expecting type procArgs", reflect.TypeOf(genericArgs))
	}
	if args.deliveryFunc == nil {
		return fmt.Errorf("unexpected nil deliver will function, should be pointing to avalid function")
	}
	if args.database == nil {
		return fmt.Errorf("unexpected nil database structure, unable to proceed")
	}
	database := args.database.Copy()
	defer database.Close()

	// find deliverable wills
	wills, err := database.GetInDelivery(time.Now())
	if err != nil {
		return err
	}
	for _, will := range wills {
		err = args.deliveryFunc(&will)
		if err != nil {
			args.errorChan <- err
			continue
		}
	}
	err = database.RemoveExausted()
	if err != nil {
		return err
	}

	return nil
}

// run the actual main routine to start looping for the
// mail dispatching service.
func run(cmd *cobra.Command, args []string) error {
	printLogo()

	// startup db
	var err error
	db, err = databaseStartup(&arguments)
	if err != nil {
		return err
	}
	defer db.Close()

	startupChans()
	// create working queue
	workingQueue = wq.NewWorkingQueue(workersize, queuesize, errc)
	// start working queue
	if err := workingQueue.Run(); err != nil {
		return err
	}
	defer workingQueue.Close()

	// run loop
	for {
		if arguments.verbose {
			log.VerboseLog("Searching routine started %s.\n", time.Now().String())
		}
		workingQueue.SendJob(processing, &procArgs{
			database:     db,
			requestTime:  time.Now(),
			errorChan:    errc,
			deliveryFunc: deliverWill,
		})
		time.Sleep(sleepingTime)
	}

	return nil
}
