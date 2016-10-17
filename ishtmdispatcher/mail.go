//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"bytes"
	"html/template"
	"sync"
)

// Internal pkgs
import (
	types "github.com/nexocrew/3nigm4/lib/commons"
)

var (
	instance  *template.Template
	once      sync.Once
	errorChan chan error
)

// getTemplate implement a singleton pattern to access
// a mail template.
func getTemplate() *template.Template {
	once.Do(func() {
		var err error
		errorChan = make(chan error, 1)
		instance, err = template.New("email").ParseFiles(arguments.htmlTemplatePath)
		if err != nil {
			errorChan <- err
		}
	})
	return instance
}

// createMailBody returns coded mail message starting
// from the will structure.
func createMailBody(content *types.Email) ([]byte, error) {
	thtml := getTemplate()
	if thtml == nil {
		return nil, <-errorChan
	}

	// https://dinosaurscode.xyz/go/2016/06/21/sending-email-using-golang/
	var buff bytes.Buffer
	err := thtml.Execute(&buff, content)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
