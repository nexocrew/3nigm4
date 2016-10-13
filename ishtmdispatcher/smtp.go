//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	_ "fmt"
	"net/smtp"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// SmtpSender the SMTP sender structure.
type SmtpSender struct {
	addr string
	port int
	auth smtp.Auth
}

// Sender interface represent sending objects.
type Sender interface {
	SendEmail(*ct.Email) error // function to actually send email messages.
}

// NewSmtpSender new Sender of type SMTP.
func NewSmtpSender(addr, usr, pwd string, port int) *SmtpSender {
	return &SmtpSender{
		addr: addr,
		port: port,
		auth: smtp.PlainAuth("", usr, pwd, addr),
	}
}

// SendEmail send a message using the defined Smtp
// inteface.
func (s *SmtpSender) SendEmail(email *ct.Email) error {
	/*
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
	*/
	return nil
}
