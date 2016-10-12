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
)

type EmailMessage struct {
	RecipientName string
	SenderName    string
	ResourceLink  string
	WillCreation  string
}

// createMailBody returns coded mail message starting
// from the will structure.
func createMailBody(content *EmailMessage) ([]byte, error) {
	tmpl, err := template.New("email").ParseFiles(args.htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	// https://dinosaurscode.xyz/go/2016/06/21/sending-email-using-golang/
	var buff bytes.buffer
	err = tmpl.Execute(&buff, content)
	if err != nil {

	}

	return nil, nil
}
