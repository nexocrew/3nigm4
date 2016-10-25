//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Internal pkgs
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	ishtmdb "github.com/nexocrew/3nigm4/lib/ishtm/db"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
	"github.com/nexocrew/3nigm4/lib/sender"
	"github.com/nexocrew/3nigm4/lib/sender/smtp"
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
	RunCmd.PersistentFlags().Uint32Var(&arguments.processScheduleMinutes, "processwait", 3, "defines the wait time for the processing routine iteration")
	RunCmd.PersistentFlags().Uint32Var(&arguments.dispatchScheduleMinutes, "dispatchtime", 5, "defines the wait time in looping for dispatching email messages produced by the processing routine")
	RunCmd.PersistentFlags().Uint32Var(&arguments.cleanupScheduleMinutes, "cleanuptime", 30, "run at defined intervals the cleanup function that remove email messages from the database")
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

// This var is used to define the used function to startup the
// sender object.
// The default, production targeting, implementation uses SMTP
// as server protocol.
var senderStartup func(*args) sender.Sender = smtpStartup

// smtpStartup SMTP sender startup function, should not be changed
// in production.
func smtpStartup(a *args) sender.Sender {
	return smtpmail.NewSmtpSender(
		a.senderAddress,
		a.senderAuthUser,
		a.senderAuthPassword,
		a.htmlTemplatePath,
		a.senderPort,
	)
}

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
	workersize           = 16
	queuesize            = 300
	minToleratedDuration = 1
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
	deliverer    sender.Sender
	criticalChan chan bool
}

// saveEmailsToDatabase save email record to the db.
func saveEmailsToDatabase(db ct.Database, w *will.Will) error {
	for _, recipient := range w.Recipients {
		email := &types.Email{
			Recipient:            recipient.Email,
			Sender:               w.Owner.Email,
			Creation:             w.Creation,
			RecipientKeyID:       recipient.KeyID,
			RecipientFingerprint: recipient.Fingerprint,
			Attachment:           w.ReferenceFile,
			DeliveryKey:          hex.EncodeToString(w.DeliveryKey),
			DeliveryDate:         time.Now().UTC(),
			Sended:               false,
		}
		err := db.SetEmail(email)
		if err != nil {
			return err
		}
	}
	return nil
}

// processEmails execute the actual async processing flow
// it is passed to the working queue and provided with all
// needed arguments.
func processEmails(genericArgs interface{}) error {
	args, ok := genericArgs.(*procArgs)
	if !ok {
		return fmt.Errorf("unexpected arguments, having %s expecting type procArgs", reflect.TypeOf(genericArgs))
	}
	if args.database == nil {
		args.criticalChan <- true
		return fmt.Errorf("unexpected nil database structure, unable to proceed")
	}
	database := args.database.Copy()
	defer database.Close()

	// find deliverable wills
	wills, err := database.GetInDelivery(time.Now())
	if err != nil {
		return err
	}
	for _, w := range wills {
		err = saveEmailsToDatabase(database, &w)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	ServiceEmail   = "3n4@nexo.cloud"
	AttachmentName = "reference.3n4"
)

// sendEmails retrieve from the db in queued emails and
// send them using a Sender service.
func sendEmails(genericArgs interface{}) error {
	args, ok := genericArgs.(*procArgs)
	if !ok {
		return fmt.Errorf("unexpected arguments, having %s expecting type procArgs", reflect.TypeOf(genericArgs))
	}
	if args.deliverer == nil {
		args.criticalChan <- true
		return fmt.Errorf("unexpected nil deliver, should be pointing to a valid struct")
	}
	if args.database == nil {
		args.criticalChan <- true
		return fmt.Errorf("unexpected nil database structure, unable to proceed")
	}
	database := args.database.Copy()
	defer database.Close()

	emails, err := database.GetEmails()
	if err != nil {
		return err
	}
	for _, email := range emails {
		err = args.deliverer.SendEmail(
			&email,
			ServiceEmail,
			fmt.Sprintf("Important data from %s", email.Sender),
			AttachmentName,
		)
		if err != nil {
			email.Sended = false
			// restore mail status by restoring sended
			// flag. Is done best effort so no error check
			// is done, otherwise all mail would be lost.
			database.SetEmail(&email)
			continue
		}
	}
	return nil
}

// cleanupSendedEmails is used to clean the database from already
// sended messages.
func cleanupSendedEmails(genericArgs interface{}) error {
	args, ok := genericArgs.(*procArgs)
	if !ok {
		return fmt.Errorf("unexpected arguments, having %s expecting type procArgs", reflect.TypeOf(genericArgs))
	}
	if args.database == nil {
		args.criticalChan <- true
		return fmt.Errorf("unexpected nil database structure, unable to proceed")
	}
	database := args.database.Copy()
	defer database.Close()

	return database.RemoveSendedEmails(time.Now())
}

func validateDuration(minutes uint32) time.Duration {
	if minutes < minToleratedDuration {
		return time.Duration(minToleratedDuration) * time.Minute
	}
	return time.Duration(minutes) * time.Minute
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

	// startup sender
	sender := senderStartup(&arguments)

	startupChans()
	// create working queue
	workingQueue = wq.NewWorkingQueue(workersize, queuesize, errc)
	// start working queue
	if err := workingQueue.Run(); err != nil {
		return err
	}
	defer workingQueue.Close()

	// timers
	processingSchedule := time.NewTicker(validateDuration(arguments.processScheduleMinutes))
	dispatchSchedule := time.NewTicker(validateDuration(arguments.dispatchScheduleMinutes))
	cleanupSchedule := time.NewTicker(validateDuration(arguments.cleanupScheduleMinutes))
	// chan used to async block schedule ops.
	critical := make(chan bool, 3)

	// run loop
	for {
		select {
		case <-processingSchedule.C:
			if arguments.verbose {
				log.VerboseLog("Deliver routine started.\n")
			}
			workingQueue.SendJob(processEmails, &procArgs{
				database:     db,
				deliverer:    sender,
				criticalChan: critical,
			})
		case <-dispatchSchedule.C:
			if arguments.verbose {
				log.VerboseLog("Dispatching routine started.\n")
			}
			workingQueue.SendJob(sendEmails, &procArgs{
				database:     db,
				deliverer:    sender,
				criticalChan: critical,
			})
		case <-cleanupSchedule.C:
			if arguments.verbose {
				log.VerboseLog("Deleting routine started.\n")
			}
			workingQueue.SendJob(cleanupSendedEmails, &procArgs{
				database:     db,
				deliverer:    sender,
				criticalChan: critical,
			})
		case <-critical:
			processingSchedule.Stop()
			dispatchSchedule.Stop()
			cleanupSchedule.Stop()
			return fmt.Errorf("timers blocked cause a critical error was produced")
		}
	}

	return nil
}
