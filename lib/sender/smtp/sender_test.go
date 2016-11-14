//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package smtpmail

// Golang std pkgs
import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
	"github.com/nexocrew/3nigm4/lib/itm"
)

func TestSmtpSending(t *testing.T) {
	once = sync.Once{}
	sender := NewSmtpSender(
		itm.S().SmtpAddress(),
		itm.S().SmtpUsername(),
		itm.S().SmtpPassword(),
		tmpFilePath,
		itm.S().SmtpPort(),
	)
	email := &ct.Email{
		Recipient:    "test@3n4.io",
		Sender:       "myuser@mail.com",
		Creation:     time.Now(),
		Attachment:   []byte("This is a fake attachment to verify final messages."),
		DeliveryKey:  "",
		DeliveryDate: time.Now(),
	}
	err := sender.SendEmail(
		email,
		"test@3n4.io",
		"test message",
		"test.txt",
	)
	if err != nil {
		t.Fatalf("Unable to send test email (template %s): %s.\n", tmpFilePath, err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"https://mailtrap.io/api/v1/inboxes/155088/messages",
		nil,
	)
	if err != nil {
		t.Fatalf("Unable to connect the mailtrap.io backend: %s.\n", err.Error())
	}
	req.Header.Set("Api-Token", itm.S().SmtpApiKey())
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform GET request, cause %s.\n", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unable to read response body, %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code having %d but expected %d: %s",
			resp.StatusCode,
			http.StatusOK,
			string(respBody),
		)
	}
	var emailsRetrieved []map[string]interface{}
	err = json.Unmarshal(respBody, &emailsRetrieved)
	if err != nil {
		t.Fatalf("Unable to unmarshal response, %s.\n", err.Error())
	}
	if len(emailsRetrieved) != 1 {
		t.Fatalf("Unexpected number of emails, having %d expecting %d.\n", len(emailsRetrieved), 1)
	}

	req, err = http.NewRequest(
		"PATCH",
		"https://mailtrap.io/api/v1/inboxes/155088/clean",
		nil,
	)
	if err != nil {
		t.Fatalf("Unable to connect the mailtrap.io backend: %s.\n", err.Error())
	}
	req.Header.Set("Api-Token", itm.S().SmtpApiKey())
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Unable to perform GET request, cause %s.\n", err.Error())
	}
	defer resp.Body.Close()
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unable to read response body, %s.\n", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code having %d but expected %d: %s",
			resp.StatusCode,
			http.StatusOK,
			string(respBody),
		)
	}

}
