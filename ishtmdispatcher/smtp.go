//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Golang std pkgs
import (
	"fmt"
	"net/mail"
	"net/smtp"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Third party pkgs
import (
	"github.com/scorredoira/email"
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
func (s *SmtpSender) SendEmail(message *ct.Email) error {
	body, err := createMailBody(message)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Important data from %s", message.Sender)
	m := email.NewHTMLMessage(subject, string(body))
	m.From = mail.Address{
		Name:    "From",
		Address: "en4@nexo.cloud",
	}
	m.To = []string{message.Recipient}
	err = m.AttachBuffer("reference.3n4", message.Attachment, false)
	if err != nil {
		return err
	}

	err = email.Send(
		fmt.Sprintf("%s:%d", s.addr, s.port),
		s.auth,
		m,
	)
	if err != nil {
		return err
	}

	return nil
}
