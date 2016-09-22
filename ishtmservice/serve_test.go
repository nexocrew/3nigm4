//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Internal dependencies.
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	ishtmct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	mockdb "github.com/nexocrew/3nigm4/lib/ishtm/mocks"
	_ "github.com/nexocrew/3nigm4/lib/ishtm/will"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

var (
	mockServiceAddress = "127.0.0.1"
	mockServicePort    = 17973
)

func TestMain(m *testing.M) {
	// start up logging facility
	log = logger.NewLogFacility("ishtmservice", true, true)

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
	authClientStartup = mockAuthStartup

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

func mockDbStartup(arguments *args) (ishtmct.Database, error) {
	mockdb := mockdb.NewMockDb(&ishtmct.DbArgs{
		Addresses: strings.Split(arguments.dbAddresses, ","),
		User:      arguments.dbUsername,
		Password:  arguments.dbPassword,
		AuthDb:    arguments.dbAuth,
	})

	log.MessageLog("Mockdb %s successfully connected.\n", arguments.dbAddresses)

	return mockdb, nil
}

func mockAuthStartup(a *args) (AuthClient, error) {
	client, err := newAuthMock()
	if err != nil {
		return nil, err
	}

	log.MessageLog("Initialised mock auth service.\n")

	return client, nil
}

func TestPing(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s:%d/v1/ping", mockServiceAddress, mockServicePort),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform request on server: %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to ping the service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}
}

func TestWillPost(t *testing.T) {
	// Login the user
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)
	if err != nil {
		t.Fatalf("Unable to marshal request body: %s.\n", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/authsession", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	// define request body
	willRequest := ct.WillPostRequest{
		Reference:      []byte("test reference data"),
		ExtensionUnit:  time.Duration(48 * time.Hour),
		NotifyDeadline: true,
		Recipients: []ct.Recipient{
			ct.Recipient{
				Name:        "recipientA",
				Email:       "recipientA@mail.com",
				KeyID:       289384,
				Fingerprint: []byte("this is a key fingerprint"),
			},
		},
	}
	body, err = json.Marshal(&willRequest)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	// create job
	req, err = http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the will POST request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will POST request on server: %s.\n", err.Error())
	}
}
