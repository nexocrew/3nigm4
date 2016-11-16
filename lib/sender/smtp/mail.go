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
	"html/template"
	"path/filepath"
	"sync"
)

// Internal pkgs
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
)

var (
	singleton *templateFactory // singleton instance;
	once      sync.Once        // concurrency safety mechanism;
)

type templateFactory struct {
	mailTemplate *template.Template // singleton pattern instance.
}

func initTemplate(templatePath string) *templateFactory {
	_, fname := filepath.Split(templatePath)
	tml, err := template.New(fname).ParseFiles(templatePath)
	if err != nil {
		return &templateFactory{}
	}
	return &templateFactory{
		mailTemplate: tml,
	}
}

func factory(templatePath string) *templateFactory {
	once.Do(func() {
		singleton = initTemplate(templatePath)
	})
	return singleton
}

// createMailBody returns coded mail message starting
// from the will structure.
func (t *templateFactory) createMailBody(content *types.Email) ([]byte, error) {
	if t.mailTemplate == nil {
		return nil, fmt.Errorf("invalid template must not be nil: an error occurred while parsing")
	}
	// https://dinosaurscode.xyz/go/2016/06/21/sending-email-using-golang/
	var buff bytes.Buffer
	err := t.mailTemplate.Execute(&buff, content)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
