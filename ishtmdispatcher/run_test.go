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
	"strings"
	"testing"
	"time"
)

// Internal pkgs
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
	ct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	mdb "github.com/nexocrew/3nigm4/lib/ishtm/mocks"
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

func TestSendingFlow(t *testing.T) {
	proc := &procArgs{
		database:     databaseInstance,
		deliverer:    senderInstance,
		criticalChan: criticalChan,
	}

	err := databaseInstance.SetEmail(&types.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	})
	if err != nil {
		t.Fatalf("Unable to add email to db: %s.\n", err.Error())
	}

	err = sendEmails(proc)
	if err != nil {
		t.Fatalf("Unable to send email message: %s.\n", err.Error())
	}
}
