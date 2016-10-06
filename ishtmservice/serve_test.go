//
// 3nigm4 ishtmservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 14/09/2016
//

package main

import (
	"bytes"
	"encoding/hex"
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
	crypto3n4 "github.com/nexocrew/3nigm4/lib/crypto"
	ishtmct "github.com/nexocrew/3nigm4/lib/ishtm/commons"
	mockdb "github.com/nexocrew/3nigm4/lib/ishtm/mocks"
	"github.com/nexocrew/3nigm4/lib/ishtm/will"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

// Third party packages
import (
	"github.com/gokyle/hotp"
)

var (
	mockServiceAddress   = "127.0.0.1"
	mockServicePort      = 17973
	GlobalEncryptionKey  = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")
	willID               string
	otp                  *hotp.HOTP
	secondaryKeys        []string
	deliveryKey          string
	lastPing             time.Time
	referenceData        = []byte("test reference data")
)

const (
	timePositiveToleranceLimit = 1 * time.Millisecond
	timeNegativeToleranceLimit = -1 * time.Millisecond
)

func decryptHotp(encryptedToken []byte) (*hotp.HOTP, error) {
	// decrypt token content
	plaintext, err := crypto3n4.AesDecrypt(
		GlobalEncryptionKey,
		encryptedToken,
		crypto3n4.CBC,
	)
	if err != nil {
		return nil, err
	}

	swtoken, err := hotp.Unmarshal(plaintext)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal token data, cause %s", err.Error())
	}
	return swtoken, nil
}

func decryptSecondaryKey(key []byte) ([]byte, error) {
	// decrypt token content
	plaintext, err := crypto3n4.AesDecrypt(
		GlobalEncryptionKey,
		key,
		crypto3n4.CBC,
	)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

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

	will.GlobalEncryptionKey = GlobalEncryptionKey
	will.GlobalEncryptionSalt = GlobalEncryptionSalt

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
	now := time.Now()
	willRequest := ct.WillPostRequest{
		Reference:      referenceData,
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

	// create will
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

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to create will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var willPostResponse ct.WillPostResponse
	err = json.Unmarshal(respBody, &willPostResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	resp.Body.Close()

	if willPostResponse.ID == "" {
		t.Fatalf("Unexpected void ID.\n")
	}
	t.Log(willPostResponse.ID)

	if willPostResponse.Credentials.QRCode == nil ||
		len(willPostResponse.Credentials.QRCode) < 2900 {
		t.Fatalf("Returned nil QRCode, not acceptable.\n")
	}
	t.Logf("QRCode png data of size: %d bytes.\n", len(willPostResponse.Credentials.QRCode))

	if len(willPostResponse.Credentials.SecondaryKeys) == 0 {
		t.Fatalf("Secondary keys are nil, should not be the case.\n")
	}
	t.Logf("Secondary keys: %v.\n", willPostResponse.Credentials.SecondaryKeys)

	// check on the db
	actualWill, err := db.GetWill(willPostResponse.ID)
	if err != nil {
		t.Fatalf("Unable to find required will: %s.\n", err.Error())
	}

	if actualWill.Deliverable != false {
		t.Fatalf("Unexpected deliverable state, having true expecting false.\n")
	}
	if actualWill.DeliveryKey == nil ||
		len(actualWill.DeliveryKey) == 0 {
		t.Fatalf("Delivery key must not be nil.\n")
	}
	t.Logf("Delivery Key: %s.\n", hex.EncodeToString(actualWill.DeliveryKey))
	if actualWill.Owner.Name != mockUserInfo.Username {
		t.Fatalf("Unexpected owner having %s expecting %s.\n", actualWill.Owner.Name, mockUserInfo.Username)
	}
	if actualWill.Owner.Email != mockUserInfo.Email {
		t.Fatalf("Unexpected owner email having %s expecting %s.\n", actualWill.Owner.Email, mockUserInfo.Email)
	}
	if len(actualWill.Owner.Credentials) != 1 {
		t.Fatalf("Unexpected credentials slice size: having %d expecting %d.\n", len(actualWill.Owner.Credentials), 1)
	}
	if actualWill.Owner.Credentials[0].SoftwareToken == nil ||
		len(actualWill.Owner.Credentials[0].SoftwareToken) == 0 {
		t.Fatalf("Unexpected credential token size having %d expecting > 0.\n", len(actualWill.Owner.Credentials[0].SoftwareToken))
	}
	if actualWill.Owner.Credentials[0].SecondaryKeys == nil ||
		len(actualWill.Owner.Credentials[0].SecondaryKeys) == 0 {
		t.Fatalf("Unexpected credential secondary key size having %d expecting > 0.\n", len(actualWill.Owner.Credentials[0].SecondaryKeys))
	}
	if actualWill.Disabled != false {
		t.Fatalf("Unexpected disable state should be false.\n")
	}
	if bytes.Compare(actualWill.ReferenceFile, willRequest.Reference) != 0 {
		t.Fatalf("Unexpected reference file.\n")
	}
	if actualWill.Creation.Sub(now) > 1*time.Millisecond {
		t.Fatalf("Unexpected creation time having %s expecting %s (diff %d ms).\n",
			actualWill.Creation.String(),
			now.String(),
			actualWill.Creation.Sub(now)/time.Millisecond)
	}
	if actualWill.LastModified.Sub(now) > 1*time.Millisecond {
		t.Fatalf("Unexpected last modified time having %s expecting %s (diff %d ms).\n",
			actualWill.LastModified.String(),
			now.String(),
			actualWill.LastModified.Sub(now)/time.Millisecond)
	}
	if actualWill.LastPing.Sub(now) > 1*time.Millisecond {
		t.Fatalf("Unexpected last ping time having %s expecting %s (diff %d ms).\n",
			actualWill.LastPing.String(),
			now.String(),
			actualWill.LastPing.Sub(now)/time.Millisecond)
	}
	// expected time should consider delivery offser if setted in the
	// settings configuration.
	expectedTtd := now.Add(willRequest.ExtensionUnit)
	if actualWill.Settings.DisableOffset != true {
		expectedTtd = expectedTtd.Add(actualWill.Settings.DeliveryOffset)
	}
	if actualWill.TimeToDelivery.Sub(expectedTtd) > timePositiveToleranceLimit ||
		actualWill.TimeToDelivery.Sub(expectedTtd) < timeNegativeToleranceLimit {
		t.Fatalf("Unexpected delivery time: having %s expecting %s (diff %d ms).\n",
			actualWill.TimeToDelivery,
			expectedTtd,
			actualWill.TimeToDelivery.Sub(expectedTtd)/time.Millisecond)
	}
	if len(actualWill.Recipients) != 1 {
		t.Fatalf("Unexpected number of recipients, having %d expecting %d.\n", len(actualWill.Recipients), 1)
	}
	if actualWill.Recipients[0].Name != willRequest.Recipients[0].Name {
		t.Fatalf("Unexpected recipient name, having %s expecting %s.\n",
			actualWill.Recipients[0].Name,
			willRequest.Recipients[0].Name)
	}
	if actualWill.Recipients[0].Email != willRequest.Recipients[0].Email {
		t.Fatalf("Unexpected recipient email, having %s expecting %s.\n",
			actualWill.Recipients[0].Email,
			willRequest.Recipients[0].Email)
	}
	if actualWill.Recipients[0].KeyID != willRequest.Recipients[0].KeyID {
		t.Fatalf("Unexpected recipient keyid, having %d expecting %d.\n",
			actualWill.Recipients[0].KeyID,
			willRequest.Recipients[0].KeyID)
	}
	if bytes.Compare(actualWill.Recipients[0].Fingerprint, willRequest.Recipients[0].Fingerprint) != 0 {
		t.Fatalf("Unexpected recipient fingerprint, having %s expecting %s.\n",
			hex.EncodeToString(actualWill.Recipients[0].Fingerprint),
			hex.EncodeToString(willRequest.Recipients[0].Fingerprint))
	}
	// set global vars for following tests
	willID = actualWill.ID
	otp, err = decryptHotp(actualWill.Owner.Credentials[0].SoftwareToken)
	if err != nil {
		t.Fatalf("Unable to decrypt with global keys the user sowftware token: %s.\n", err.Error())
	}
	secondaryKeys = willPostResponse.Credentials.SecondaryKeys
	deliveryKey = hex.EncodeToString(actualWill.DeliveryKey)
}

func TestWillPatchWithOtp(t *testing.T) {
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

	if otp == nil {
		t.Fatalf("Otp must not be nil.\n")
	}

	time.Sleep(300 * time.Millisecond)
	// define request body
	willRequest := ct.WillPatchRequest{
		Otp: otp.OTP(),
	}
	body, err = json.Marshal(&willRequest)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	now := time.Now()
	// patch will
	req, err = http.NewRequest(
		"PATCH",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s", mockServiceAddress, mockServicePort, willID),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the will PATCH request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will PATCH request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to patch will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var stdResponse ct.StandardResponse
	err = json.Unmarshal(respBody, &stdResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	resp.Body.Close()

	// check ping value
	actualWill, err := db.GetWill(willID)
	if err != nil {
		t.Fatalf("Unable to find required will: %s.\n", err.Error())
	}

	// expected time should consider delivery offser if setted in the
	// settings configuration.
	expectedTtd := now.Add(actualWill.Settings.ExtensionUnit)
	if actualWill.Settings.DisableOffset != true {
		expectedTtd = expectedTtd.Add(actualWill.Settings.DeliveryOffset)
	}
	if actualWill.TimeToDelivery.Sub(expectedTtd) > timePositiveToleranceLimit ||
		actualWill.TimeToDelivery.Sub(expectedTtd) < timeNegativeToleranceLimit {
		t.Fatalf("Unexpected delivery time: having %s expecting %s (diff %d ms) with extension %d ms.\n",
			actualWill.TimeToDelivery,
			expectedTtd,
			actualWill.TimeToDelivery.Sub(expectedTtd)/time.Millisecond,
			actualWill.Settings.ExtensionUnit/time.Millisecond)
	}

	time.Sleep(700 * time.Millisecond)
	// define request body
	willRequest = ct.WillPatchRequest{
		Otp: otp.OTP(),
	}
	body, err = json.Marshal(&willRequest)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	now = time.Now()
	// patch will
	req, err = http.NewRequest(
		"PATCH",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s", mockServiceAddress, mockServicePort, willID),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the will PATCH request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will PATCH request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to patch will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	// check ping value
	actualWill, err = db.GetWill(willID)
	if err != nil {
		t.Fatalf("Unable to find required will: %s.\n", err.Error())
	}

	// expected time should consider delivery offser if setted in the
	// settings configuration.
	expectedTtd = now.Add(actualWill.Settings.ExtensionUnit)
	if actualWill.Settings.DisableOffset != true {
		expectedTtd = expectedTtd.Add(actualWill.Settings.DeliveryOffset)
	}
	if actualWill.TimeToDelivery.Sub(expectedTtd) > timePositiveToleranceLimit ||
		actualWill.TimeToDelivery.Sub(expectedTtd) < timeNegativeToleranceLimit {
		t.Fatalf("Unexpected delivery time: having %s expecting %s (diff %d ms) with extension %d ms.\n",
			actualWill.TimeToDelivery,
			expectedTtd,
			actualWill.TimeToDelivery.Sub(expectedTtd)/time.Millisecond,
			actualWill.Settings.ExtensionUnit/time.Millisecond)
	}
}

func TestWillPatchWithSecondary(t *testing.T) {
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

	time.Sleep(300 * time.Millisecond)
	if len(secondaryKeys) < 1 {
		t.Fatalf("Unable to proceed without at least a secondary key.\n")
	}
	// define request body
	willRequest := ct.WillPatchRequest{
		SecondaryKey: secondaryKeys[0],
	}
	body, err = json.Marshal(&willRequest)
	if err != nil {
		t.Fatalf("Unable to marshal request bodu: %s.\n", err.Error())
	}

	now := time.Now()
	// patch will
	req, err = http.NewRequest(
		"PATCH",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s", mockServiceAddress, mockServicePort, willID),
		bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Unable to prepare the will PATCH request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will PATCH request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to patch will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var stdResponse ct.StandardResponse
	err = json.Unmarshal(respBody, &stdResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	resp.Body.Close()

	// check ping value
	actualWill, err := db.GetWill(willID)
	if err != nil {
		t.Fatalf("Unable to find required will: %s.\n", err.Error())
	}

	// expected time should consider delivery offser if setted in the
	// settings configuration.
	expectedTtd := now.Add(actualWill.Settings.ExtensionUnit)
	if actualWill.Settings.DisableOffset != true {
		expectedTtd = expectedTtd.Add(actualWill.Settings.DeliveryOffset)
	}
	if actualWill.TimeToDelivery.Sub(expectedTtd) > timePositiveToleranceLimit ||
		actualWill.TimeToDelivery.Sub(expectedTtd) < timeNegativeToleranceLimit {
		t.Fatalf("Unexpected delivery time: having %s expecting %s (diff %d ms) with extension %d ms.\n",
			actualWill.TimeToDelivery,
			expectedTtd,
			actualWill.TimeToDelivery.Sub(expectedTtd)/time.Millisecond,
			actualWill.Settings.ExtensionUnit/time.Millisecond)
	}
	lastPing = now
}

func TestWillGetForOwner(t *testing.T) {
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

	// get will
	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s", mockServiceAddress, mockServicePort, willID),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the will GET request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will GET request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to get will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var getResponse ct.WillGetResponse
	err = json.Unmarshal(respBody, &getResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	resp.Body.Close()

	if getResponse.ID != willID {
		t.Fatalf("Unexpected will id, having %s expecting %s.\n", getResponse.ID, willID)
	}
	if len(getResponse.Recipients) == 0 {
		t.Fatalf("Owner must be able to retrieve recipients.\n")
	}
	if getResponse.ExtensionUnit == 0 {
		t.Fatalf("Owner must be able to retrieve extension unit.\n")
	}
	if getResponse.NotifyDeadline != true {
		t.Fatalf("Owner must retrieve notify deadline flag.\n")
	}
	if getResponse.LastPing.Sub(lastPing) > timePositiveToleranceLimit ||
		getResponse.LastPing.Sub(lastPing) < timeNegativeToleranceLimit {
		t.Fatalf("Last ping time is wrong should be ~0 but found %d ms.\n",
			getResponse.LastPing.Sub(lastPing)/1*time.Millisecond)
	}
	if bytes.Compare(referenceData, getResponse.ReferenceFile) != 0 {
		t.Fatalf("Unexpected reference file.\n")
	}
}

func TestWillGetForRecipient(t *testing.T) {
	client := &http.Client{}
	// get will
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s?deliverykey=%s",
			mockServiceAddress,
			mockServicePort,
			willID,
			deliveryKey),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the will GET request: %s.\n", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will GET request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to get will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var getResponse ct.WillGetResponse
	err = json.Unmarshal(respBody, &getResponse)
	if err != nil {
		t.Fatalf("Unable to unmarshal response body: %s.\n", err.Error())
	}
	resp.Body.Close()

	if getResponse.ID != willID {
		t.Fatalf("Unexpected will id, having %s expecting %s.\n", getResponse.ID, willID)
	}
	if len(getResponse.Recipients) != 0 {
		t.Fatalf("Recipient must not be able to retrieve recipients.\n")
	}
	if getResponse.ExtensionUnit != 0 {
		t.Fatalf("Recipient must not be able to retrieve extension unit.\n")
	}
	if getResponse.NotifyDeadline == true {
		t.Fatalf("Recipient must not retrieve notify deadline flag.\n")
	}
	if getResponse.LastPing.Sub(lastPing) > timePositiveToleranceLimit ||
		getResponse.LastPing.Sub(lastPing) < timeNegativeToleranceLimit {
		t.Fatalf("Last ping time is wrong should be ~0 but found %d ms.\n",
			getResponse.LastPing.Sub(lastPing)/1*time.Millisecond)
	}
	if bytes.Compare(referenceData, getResponse.ReferenceFile) != 0 {
		t.Fatalf("Unexpected reference file.\n")
	}
}

func TestWillDelete(t *testing.T) {
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

	// backup it
	actualWill, err := db.GetWill(willID)
	if err != nil {
		t.Fatalf("Unable to find required will: %s.\n", err.Error())
	}

	if otp == nil {
		t.Fatalf("Otp must not be nil.\n")
	}

	// delete will with otp
	req, err = http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s?otp=%s",
			mockServiceAddress,
			mockServicePort,
			willID,
			otp.OTP()),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the will DELETE request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will DELETE request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to delete will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}

	// restore will
	err = db.SetWill(actualWill)
	if err != nil {
		t.Fatalf("Unable to restore will in the database: %s.\n", err.Error())
	}

	if len(secondaryKeys) < 2 {
		t.Fatalf("Unable to proceed without at least a secondary key.\n")
	}

	// delete will with secondary key invalid
	req, err = http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s?secondarykey=%s",
			mockServiceAddress,
			mockServicePort,
			willID,
			secondaryKeys[0]),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the will DELETE request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will DELETE request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Try delating with already used secondary should return %d but produced %d.\n", http.StatusUnauthorized, resp.StatusCode)
	}

	// delete will with secondary key
	req, err = http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://%s:%d/v1/ishtm/will/%s?secondarykey=%s",
			mockServiceAddress,
			mockServicePort,
			willID,
			secondaryKeys[1]),
		nil)
	if err != nil {
		t.Fatalf("Unable to prepare the will DELETE request: %s.\n", err.Error())
	}
	req.Header.Set(ct.SecurityTokenKey, session.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform will DELETE request on server: %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to delete will returned %d but expected %d.\n", resp.StatusCode, http.StatusOK)
	}
}
