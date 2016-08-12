//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
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
	log = logger.NewLogFacility("storageservice", true, true)

	arguments = args{
		verbose:            true,
		colored:            true,
		dbAddresses:        fmt.Sprintf("%s:%d", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:         itm.S().DbUserName(),
		dbPassword:         itm.S().DbPassword(),
		dbAuth:             itm.S().DbAuth(),
		address:            mockServiceAddress,
		port:               mockServicePort,
		s3Endpoint:         itm.S().S3Endpoint(),
		s3Region:           itm.S().S3Region(),
		s3Id:               itm.S().S3Id(),
		s3Secret:           itm.S().S3Secret(),
		s3Token:            itm.S().S3Token(),
		s3Bucket:           itm.S().S3Bucket(),
		s3QueueSize:        200,
		s3WorkingQueueSize: 12,
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

func mockDbStartup(arguments *args) (database, error) {
	mockdb := newMockDb(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
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
		t.Fatalf("Unable to ping the service, returned %d but expected 200.\n", resp.StatusCode)
	}
}

func TestLoginAndLogout(t *testing.T) {
	loginBody := ct.LoginRequest{
		Username: mockUserInfo.Username,
		Password: mockUserPassword,
	}
	body, err := json.Marshal(&loginBody)

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s:%d/v1/login", mockServiceAddress, mockServicePort),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the login request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform login request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to access login service, returned %d but expected 200.\n", resp.StatusCode)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var session ct.LoginResponse
	err = json.Unmarshal(respBody, &session)
	if session.Token == "" ||
		len(session.Token) == 0 {
		t.Fatalf("Invalid token: should not be nil.\n")
	}
	resp.Body.Close()

	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s:%d/v1/logout", mockServiceAddress, mockServicePort),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the logout request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform logout request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to perform logout, returned %d but expected 200.\n", resp.StatusCode)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var invalidated ct.LogoutResponse

	err = json.Unmarshal(respBody, &invalidated)
	if invalidated.Invalidated == "" ||
		len(invalidated.Invalidated) == 0 {
		t.Fatalf("Invalid invalidated field: should not be nil.\n")
	}
	resp.Body.Close()

	if session.Token != invalidated.Invalidated {
		t.Fatalf("Expecting invalidating %s but found %s.\n", session.Token, invalidated.Invalidated)
	}
}
