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
	"os"
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
)

// Third party pkgs
import (
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2/bson"
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
	Example: "ishtmdispatcher run -d 127.0.0.1:27017 -u dbuser -w dbpwd --smtpaddress 192.168.0.1 --smtpport 443 --smtpuser username --smtppwd pwd --template /home/user/template.html -v",
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
	RunCmd.PersistentFlags().StringVarP(&arguments.senderEmailAddress, "senderemail", "m", "ishtm@3n4.io", "the email address to be used as sender in email messages")
	RunCmd.PersistentFlags().StringVarP(&arguments.htmlTemplatePath, "template", "", "", "specify the email template to be used for notification")
	RunCmd.PersistentFlags().Uint32VarP(&arguments.processScheduleMinutes, "processwait", "", 3, "defines the wait time for the processing routine iteration in minutes")
	RunCmd.PersistentFlags().Uint32VarP(&arguments.dispatchScheduleMinutes, "dispatchtime", "", 5, "defines the wait time in looping for dispatching email messages produced by the processing routine in minutes")
	RunCmd.PersistentFlags().Uint32VarP(&arguments.cleanupScheduleMinutes, "cleanuptime", "", 30, "run at defined intervals the cleanup function that remove email messages from the database in minutes")
	RunCmd.PersistentFlags().BoolVarP(&arguments.now, "now", "", false, "for debugging execute routine every 10 seconds")
	// files parameters
	RunCmd.RunE = run
}

// Global database referring variable to be copied and released by
// each goroutine.
var db ct.Database

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
	minToleratedDuration = 1
)

// procArgs processing func used arguments.
type procArgs struct {
	database     ct.Database
	deliverer    sender.Sender
	senderEmail  string
	criticalChan chan bool
	errorChan    chan error
}

type sendingArgs struct {
	message      types.Email
	database     ct.Database
	deliverer    sender.Sender
	senderEmail  string
	criticalChan chan bool
	errorChan    chan error
}

// saveEmailsToDatabase save email record to the db.
func saveEmailsToDatabase(db ct.Database, w *will.Will) error {
	for _, recipient := range w.Recipients {
		email := &types.Email{
			ObjectID:             bson.NewObjectId(),
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
	if arguments.verbose {
		log.VerboseLog("Processing routine triggered: %s.\n", time.Now().UTC().String())
	}

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
	AttachmentName = "reference.3n4"
)

// sendingAsync working queue ready function to actually
// send messages using a Sender interface.
func sendingAsync(genericArgs interface{}) error {
	args, ok := genericArgs.(*sendingArgs)
	if !ok {
		return fmt.Errorf("unexpected arguments, having %s expecting type sendingArgs", reflect.TypeOf(genericArgs))
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

	err := args.deliverer.SendEmail(
		&args.message,
		args.senderEmail,
		fmt.Sprintf("Important data from %s", args.message.Sender),
		AttachmentName,
	)
	if err != nil {
		args.errorChan <- fmt.Errorf("error sending email: %s", err.Error())
		args.message.Sended = false
		// restore mail status by restoring sended
		// flag.
		err = database.SetEmail(&args.message)
		if err != nil {
			return err
		}
	}
	return nil
}

// sendEmails retrieve from the db in queued emails and
// send them using a Sender service.
func sendEmails(genericArgs interface{}) error {
	if arguments.verbose {
		log.VerboseLog("Sending routine triggered: %s.\n", time.Now().UTC().String())
	}

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
		// mail sending is done using the workingqueue
		// to distribute processing pressure.
		workingQueue.SendJob(sendingAsync, &sendingArgs{
			message:      email,
			deliverer:    args.deliverer,
			senderEmail:  args.senderEmail,
			database:     args.database,
			errorChan:    args.errorChan,
			criticalChan: args.criticalChan,
		})
	}
	return nil
}

// cleanupSendedEmails is used to clean the database from already
// sended messages.
func cleanupSendedEmails(genericArgs interface{}) error {
	if arguments.verbose {
		log.VerboseLog("Cleanup routine triggered: %s.\n", time.Now().UTC().String())
	}

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

	if arguments.htmlTemplatePath == "" {
		return fmt.Errorf("unable to access mail template, unable to start dispatching routine")
	}
	_, err = os.Stat(arguments.htmlTemplatePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("template file do not exist")
	}

	// startup sender
	sender := senderStartup(&arguments)

	// timers
	var processingSchedule, dispatchSchedule, cleanupSchedule *time.Ticker
	// chan used to async block schedule ops.
	critical := make(chan bool, 3)
	if arguments.now {
		workingQueue.SendJob(processEmails, &procArgs{
			database:     db,
			deliverer:    sender,
			senderEmail:  arguments.senderEmailAddress,
			criticalChan: critical,
			errorChan:    errc,
		})
		time.Sleep(1 * time.Second)
		workingQueue.SendJob(sendEmails, &procArgs{
			database:     db,
			deliverer:    sender,
			senderEmail:  arguments.senderEmailAddress,
			criticalChan: critical,
			errorChan:    errc,
		})
		time.Sleep(1 * time.Second)
		workingQueue.SendJob(cleanupSendedEmails, &procArgs{
			database:     db,
			deliverer:    sender,
			senderEmail:  arguments.senderEmailAddress,
			criticalChan: critical,
			errorChan:    errc,
		})
		log.MessageLog("Completed jobs with critical signals: %d.\n", len(critical))
	} else {
		processingSchedule = time.NewTicker(validateDuration(arguments.processScheduleMinutes))
		dispatchSchedule = time.NewTicker(validateDuration(arguments.dispatchScheduleMinutes))
		cleanupSchedule = time.NewTicker(validateDuration(arguments.cleanupScheduleMinutes))

		// run loop
		for {
			select {
			case <-processingSchedule.C:
				workingQueue.SendJob(processEmails, &procArgs{
					database:     db,
					deliverer:    sender,
					senderEmail:  arguments.senderEmailAddress,
					criticalChan: critical,
					errorChan:    errc,
				})
			case <-dispatchSchedule.C:
				workingQueue.SendJob(sendEmails, &procArgs{
					database:     db,
					deliverer:    sender,
					senderEmail:  arguments.senderEmailAddress,
					criticalChan: critical,
					errorChan:    errc,
				})
			case <-cleanupSchedule.C:
				workingQueue.SendJob(cleanupSendedEmails, &procArgs{
					database:     db,
					deliverer:    sender,
					senderEmail:  arguments.senderEmailAddress,
					criticalChan: critical,
					errorChan:    errc,
				})
			case <-critical:
				processingSchedule.Stop()
				dispatchSchedule.Stop()
				cleanupSchedule.Stop()
				return fmt.Errorf("timers blocked cause a critical error was produced")
			}
		}
	}

	return nil
}
