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
	"reflect"
	"strings"
	"testing"
	"time"
)

// Internal pkgs
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	mdb "github.com/nexocrew/3nigm4/lib/ishtm/mocks"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/sender"
	"github.com/nexocrew/3nigm4/lib/sender/mock"
)

func mocksenderStartup(a *args) sender.Sender {
	return sendermock.NewMockSender()
}

func mockDbStartup(arguments *args) (ct.Database, error) {
	mockdb := mdb.NewMockDb(&ct.DbArgs{
		Addresses: strings.Split(arguments.dbAddresses, ","),
		User:      arguments.dbUsername,
		Password:  arguments.dbPassword,
		AuthDb:    arguments.dbAuth,
	})

	log.MessageLog("Mockdb %s successfully connected.\n", arguments.dbAddresses)
	return mockdb, nil
}

var (
	databaseInstance ct.Database
	senderInstance   sender.Sender
	criticalChan     chan bool
)

func TestMain(m *testing.M) {
	// start up logging facility
	log = logger.NewLogFacility("ishtmdispatcher", true, true)

	arguments = args{
		verbose:            true,
		colored:            true,
		dbAddresses:        fmt.Sprintf("%s:%d", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:         itm.S().DbUserName(),
		dbPassword:         itm.S().DbPassword(),
		dbAuth:             itm.S().DbAuth(),
		senderPort:         25,
		senderAuthUser:     "user",
		senderAuthPassword: "password",
		// schedulers
		processScheduleMinutes:  1,
		dispatchScheduleMinutes: 2,
		cleanupScheduleMinutes:  2,
	}
	databaseStartup = mockDbStartup
	senderStartup = mocksenderStartup

	var err error
	databaseInstance, err = databaseStartup(&arguments)
	if err != nil {
		log.CriticalLog("Unable to start mock database:%s.\n", err.Error())
		os.Exit(1)
	}
	senderInstance = senderStartup(&arguments)
	criticalChan = make(chan bool, 3)

	os.Exit(m.Run())
}

var (
	referenceMail = &types.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
		Sended:               false,
	}
)

func TestSendingFlow(t *testing.T) {
	proc := &procArgs{
		database:     databaseInstance,
		deliverer:    senderInstance,
		criticalChan: criticalChan,
	}

	err := databaseInstance.SetEmail(referenceMail)
	if err != nil {
		t.Fatalf("Unable to add email to db: %s.\n", err.Error())
	}

	emails, err := databaseInstance.GetEmails()
	if err != nil {
		t.Fatalf("Unable to retrieve emails: %s.\n", err.Error())
	}
	if len(emails) != 1 {
		t.Fatalf("Should have 1 email in queue but found %d.\n", len(emails))
	}

	err = databaseInstance.SetEmail(referenceMail)
	if err != nil {
		t.Fatalf("Unable to add email to db: %s.\n", err.Error())
	}

	err = sendEmails(proc)
	if err != nil {
		t.Fatalf("Unable to send email message: %s.\n", err.Error())
	}

	emails, err = databaseInstance.GetEmails()
	if err != nil {
		t.Fatalf("Unable to retrieve emails: %s.\n", err.Error())
	}
	if len(emails) != 0 {
		t.Fatalf("Should have no email in queue but found %d.\n", len(emails))
	}

	mockSender, ok := senderInstance.(*sendermock.MockSender)
	if !ok {
		t.Fatalf("Unexpected type of sender, having %s expecting MockSender.\n", reflect.TypeOf(senderInstance))
	}
	if len(mockSender.Sended) != 1 {
		t.Fatalf("Unexpected count of sended email: having %d expecting 1 %v.\n", len(mockSender.Sended), senderInstance)
	}

	for _, v := range mockSender.Sended {
		if v.FromAddress != ServiceEmail {
			t.Fatalf("Unexpected sender: having %s expecting %s.\n", v.FromAddress, ServiceEmail)
		}
		if v.AttachmentName != AttachmentName {
			t.Fatalf("Unexpected attachment: having %s expecting %s.\n", v.AttachmentName, AttachmentName)
		}
		if reflect.DeepEqual(v.Email, referenceMail) != true {
			t.Fatalf("Unexpected email: different from reference content.\n")
		}
		referenceSubject := fmt.Sprintf("Important data from %s", v.Email.Sender)
		if v.Subject != referenceSubject {
			t.Fatalf("Unexpected subject: having %s expecting %s.\n", v.Subject, referenceSubject)
		}
	}
	// cleanup
	databaseInstance = nil
	databaseInstance, err = databaseStartup(&arguments)
	if err != nil {
		t.Fatalf("Unable to start mock database:%s.\n", err.Error())
	}
}

func createTestWill(t *testing.T) *will.Will {
	owner := &will.OwnerID{
		Name:  "userA",
		Email: "userA@mail.com",
	}
	settings := &will.Settings{
		ExtensionUnit:  time.Duration(3 * time.Millisecond),
		DisableOffset:  true,
		NotifyDeadline: true,
		DeliveryOffset: time.Duration(3 * time.Millisecond),
	}
	recipients := []types.Recipient{
		types.Recipient{
			Email: "recipientA@mail.com",
			Name:  "Recipient A",
		},
	}

	will.GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	will.GlobalEncryptionSalt = []byte("thisissa")

	w, _, err := will.NewWill(owner, []byte("This is a mock reference file"), settings, recipients)
	if err != nil {
		t.Fatalf("Unable to create will instance: %s.\n", err.Error())
	}
	w.TimeToDelivery = time.Now().UTC()

	return w
}

func TestProcessingFlow(t *testing.T) {
	proc := &procArgs{
		database:     databaseInstance,
		deliverer:    senderInstance,
		criticalChan: criticalChan,
	}
	w := createTestWill(t)
	err := databaseInstance.SetWill(w)
	if err != nil {
		t.Fatalf("Unable to add will: %s.\n", err.Error())
	}
	time.Sleep(1 * time.Second)

	err = processEmails(proc)
	if err != nil {
		t.Fatalf("Unable to process will to produce messages: %s.\n", err.Error())
	}
	emails, err := databaseInstance.GetEmails()
	if err != nil {
		t.Fatalf("Unable to find emails: %s.\n", err.Error())
	}
	if len(emails) != 1 {
		t.Fatalf("Unexpected number of emails, having %d expecting %d.\n", len(emails), 1)
	}
	selected := &emails[0]
	if selected.Sender != w.Owner.Email {
		t.Fatalf("Unexpected sender, having %s expecting %s.\n", selected.Sender, w.Owner.Email)
	}
	if len(w.Recipients) != 1 {
		t.Fatalf("Unexpected number of recipients, having %d expecting %d.\n", len(w.Recipients), 1)
	}
	if selected.Recipient != w.Recipients[0].Email {
		t.Fatalf("Unexpected recipient, having %s expecting %s.\n", selected.Recipient, w.Recipients[0].Email)
	}
}
