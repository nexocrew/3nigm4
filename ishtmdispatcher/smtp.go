//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"net/smtp"
)

// Internal pkgs
import (
	wl "github.com/nexocrew/3nigm4/lib/ishtm/will"
)

// SmtpSender the SMTP sender structure.
type SmtpSender struct {
	addr string
	port int
	auth smtp.Auth
}

// Sender interface represent sending objects.
type Sender interface {
	SendWill(*wl.Will) error // function to actually send will messages.
}

// NewSmtpSender new Sender of type SMTP.
func NewSmtpSender(addr, usr, pwd string, port int) *SmtpSender {
	return &SmtpSender{
		addr: addr,
		port: port,
		auth: smtp.PlainAuth("", usr, pwd, addr),
	}
}

// SendWill send a message using the defined Smtp
// inteface.
func (s *SmtpSender) SendWill(will *wl.Will) error {
	errorDescription := make([]string, 0)
	unsentMessages := make([][]byte, 0)
	for _, recipient := range will.Recipients {
		body, err := createMailBody(recipient, will)
		if err != nil {
			errorDescription = append(errorDescription, err.Error())
			continue
		}
		err = smtp.SendMail(
			fmt.Sprintf("%s:%d", s.addr, s.port),
			s.auth,
			"en4@nexo.cloud",
			[]string{recipient},
			body,
		)
		if err != nil {
			errorDescription = append(errorDescription, err.Error())
			s.unsentMessages = append(s.unsentMessages, body)
			continue
		}
	}
	if len(unsentMessages) != 0 &&
		db != nil {
		err := db.StoreUnsentMessages(unsentMessages)
		if err != nil {
			return err
		}
	}
	if len(errorDescription) != 0 {
		return fmt.Errorf("Founded %d errors while proceeding sending messages: %v",
			len(errorDescription),
			errorDescription,
		)
	}
	return nil
}
