//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package smtpmail

// Golang std pkgs
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

const (
	mailTemplateData = `
	<!DOCTYPE html>
	<html>
	<h1>Hello {{ .Recipient }}</h1>
	<p>We are sending you a file made available from {{ .Sender }} in case something happened to him or her on date {{ .Creation }}, please follow the following instruction to access the shared resources&#58;</p>
	<h3>Scenario A&#58;</h3>
	<ul>
	<li>Download 3n4cli from 3n4.io and install it (see instructions for details)&#59;</li>
	<li>Download and save locally the file attached to this email as for example resources.3rf&#59;</li>
	<li>Use the command <code>3n4cli store download -M -o /tmp/file.ext -r /home/user/resources.3rf</code>&#59;</li>
	<li>Download the actual content&#46;</li>
	</ul>
	<h3>Scenario B&#58;</h3>
	<ul>
	<li>Download 3n4cli from 3n4.io and install it (see instructions for details)&#59;</li>
	<li>Download the attached file using command <code>3n4cli ishtm get --id {{ .DeliveryKey }} --output /home/user/reference.3n4</code>&#59;</li>
	<li>Use the command <code>3n4cli store download -M -o /tmp/file.ext -r /home/user/resources.3rf</code>&#59;</li>
	<li>Download the actual content&#46;</li>
	</ul>
	</html>
	`
	wrongMailTemplateData = `
	{{ template "email" }} {
	<h1>Hello {{ .Recipient }}</h1>
	<p>We are sending you a file made available from {{ .Sender }} in case something happened to him or her on date {{ .Creation }}, please follow the following instruction to access the shared resources&#58;</p>
	<h3>Scenario A&#58;</h3>
	<ul>
	<li>Download 3n4cli from 3n4.io and install it (see instructions for details)&#59;</li>
	<li>Download and save locally the file attached to this email as for example resources.3rf&#59;</li>
	<li>Use the command <code>3n4cli store download -M -o /tmp/file.ext -r /home/user/resources.3rf</code>&#59;</li>
	<li>Download the actual conte
	</ul>
	<h3>Scenario B&#58;</h3>
	{{ end }}
	`
	referenceHtmlData = `
	<!DOCTYPE html>
	<html>
	<h1>Hello %s</h1>
	<p>We are sending you a file made available from %s in case something happened to him or her on date %s, please follow the following instruction to access the shared resources&#58;</p>
	<h3>Scenario A&#58;</h3>
	<ul>
	<li>Download 3n4cli from 3n4.io and install it (see instructions for details)&#59;</li>
	<li>Download and save locally the file attached to this email as for example resources.3rf&#59;</li>
	<li>Use the command <code>3n4cli store download -M -o /tmp/file.ext -r /home/user/resources.3rf</code>&#59;</li>
	<li>Download the actual content&#46;</li>
	</ul>
	<h3>Scenario B&#58;</h3>
	<ul>
	<li>Download 3n4cli from 3n4.io and install it (see instructions for details)&#59;</li>
	<li>Download the attached file using command <code>3n4cli ishtm get --id %s --output /home/user/reference.3n4</code>&#59;</li>
	<li>Use the command <code>3n4cli store download -M -o /tmp/file.ext -r /home/user/resources.3rf</code>&#59;</li>
	<li>Download the actual content&#46;</li>
	</ul>
	</html>
	`
)

var (
	tmpFilePath   string
	wrongFilePath string
)

func TestMain(m *testing.M) {
	var err error
	tmpd, err := ioutil.TempDir("", "3n4ishtmdispatch")
	if err != nil {
		fmt.Printf("Unable to create tmp file: %s.\n", err.Error())
		os.Exit(1)
	}
	defer os.Remove(tmpd)

	tmpFilePath = path.Join(tmpd, "template.hmtl")

	err = ioutil.WriteFile(tmpFilePath, []byte(mailTemplateData), 0700)
	if err != nil {
		fmt.Printf("Unable to write data to file %s, cause %s.\n", tmpFilePath, err.Error())
		os.Exit(1)
	}

	tmpdwrong, err := ioutil.TempDir("", "3n4ishtmdispatchwrong")
	if err != nil {
		fmt.Printf("Unable to create tmp file: %s.\n", err.Error())
		os.Exit(1)
	}
	defer os.Remove(tmpdwrong)

	wrongFilePath = path.Join(tmpdwrong, "wrong.hmtl")

	err = ioutil.WriteFile(wrongFilePath, []byte(wrongMailTemplateData), 0700)
	if err != nil {
		fmt.Printf("Unable to write data to file %s, cause %s.\n", wrongFilePath, err.Error())
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestTemplateExecution(t *testing.T) {
	once = sync.Once{}
	email := &ct.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	}
	data, err := factory(tmpFilePath).createMailBody(email)
	if err != nil {
		t.Fatalf("Unable to create mail body: %s.\n", err.Error())
	}
	reference := fmt.Sprintf(referenceHtmlData, email.Recipient, email.Sender, email.Creation.String(), email.DeliveryKey)
	reference = strings.Replace(reference, "+", "&#43;", -1)
	if bytes.Compare(data, []byte(reference)) != 0 {
		t.Fatalf("Unexpected body produced, having %s expecting %s.\n", reference, string(data))
	}
}

func TestWrongTemplateExecution(t *testing.T) {
	once = sync.Once{}
	email := &ct.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	}
	data, err := factory(wrongFilePath).createMailBody(email)
	if data != nil {
		t.Fatalf("Parsing erroneous html template must return nil data")
	}
	if err == nil {
		t.Fatalf("Parsing erroneous html template must return an error")
	}
}

func TestExecuteOnMultipleIterations(t *testing.T) {
	once = sync.Once{}
	email := &ct.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	}
	reference := fmt.Sprintf(referenceHtmlData, email.Recipient, email.Sender, email.Creation.String(), email.DeliveryKey)
	reference = strings.Replace(reference, "+", "&#43;", -1)

	var wg sync.WaitGroup
	errorChan := make(chan error, 1000)
	for idx := 0; idx < 1000; idx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := factory(tmpFilePath).createMailBody(email)
			if err != nil {
				errorChan <- err
				return
			}
			if bytes.Compare(data, []byte(reference)) != 0 {
				errorChan <- fmt.Errorf("invalid produced html data, not equal to reference")
			}
			return
		}()
		time.Sleep(100 * time.Nanosecond)
	}
	wg.Wait()

	if len(errorChan) != 0 {
		t.Fatalf("Parallel template creation produced errors (%v).\n", errorChan)
	}
}

func TestWrongExecuteOnMultipleIterations(t *testing.T) {
	once = sync.Once{}
	email := &ct.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	}
	reference := fmt.Sprintf(referenceHtmlData, email.Recipient, email.Sender, email.Creation.String(), email.DeliveryKey)
	reference = strings.Replace(reference, "+", "&#43;", -1)

	var wg sync.WaitGroup
	errorChan := make(chan error, 1000)
	for idx := 0; idx < 1000; idx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := factory(wrongFilePath).createMailBody(email)
			if err != nil {
				errorChan <- err
				return
			}
			if bytes.Compare(data, []byte(reference)) != 0 {
				errorChan <- fmt.Errorf("invalid produced html data, not equal to reference")
			}
			return
		}()
		time.Sleep(100 * time.Nanosecond)
	}
	wg.Wait()

	if len(errorChan) != 1000 {
		t.Fatalf("Parallel template creation produced few errrors, should be %d but had %d.\n", 1000, len(errorChan))
	}
}
