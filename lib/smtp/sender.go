//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

// Package smtpmail implements a sender compliant struct
// managing SMTP mail servers.
package smtpmail

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
	addr         string
	port         int
	auth         smtp.Auth
	templatePath string
}

// Sender interface represent sending objects.
type Sender interface {
	SendEmail(*ct.Email, string, string, string) error // function to actually send email messages.
}

// NewSmtpSender new Sender of type SMTP.
func NewSmtpSender(addr, usr, pwd, template string, port int) *SmtpSender {
	return &SmtpSender{
		addr:         addr,
		port:         port,
		auth:         smtp.PlainAuth("", usr, pwd, addr),
		templatePath: template,
	}
}

// SendEmail send a message using the defined Smtp
// inteface.
func (s *SmtpSender) SendEmail(content *ct.Email, fromAddress, subject, attachmentName string) error {
	body, err := createMailBody(content, s.templatePath)
	if err != nil {
		return err
	}
	m := email.NewHTMLMessage(subject, string(body))
	m.From = mail.Address{
		Name:    "From",
		Address: fromAddress,
	}
	m.To = []string{content.Recipient}
	err = m.AttachBuffer(attachmentName, content.Attachment, false)
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
