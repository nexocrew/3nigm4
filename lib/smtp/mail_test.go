//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package smtpmail

// Golang std pkgs
import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
)

var (
	tmpFilePath string
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

	os.Exit(m.Run())
}

func TestTemplateExecution(t *testing.T) {
	data, err := createMailBody(&ct.Email{
		Recipient:            "userB",
		Sender:               "userA",
		Creation:             time.Now(),
		RecipientKeyID:       2534,
		RecipientFingerprint: []byte("a38eyr3ye72t6e3"),
		DeliveryKey:          "01234567890",
		Attachment:           []byte("This is a fake test attachment"),
	}, tmpFilePath)
	if err != nil {
		t.Fatalf("Unable to create mail body: %s.\n", err.Error())
	}
	t.Logf("Data %s\n", string(data))

}
