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

type SmtpSender struct {
	addr string
	port int
	auth smtp.Auth
}

type Sender interface {
	SendWill(*wl.Will) error
}

func NewSmtpSender(addr, usr, pwd string, port int) *SmtpSender {
	return &SmtpSender{
		addr: addr,
		port: port,
		auth: smtp.PlainAuth("", usr, pwd, addr),
	}
}

func createMailBody(will *wl.Will) ([]byte, error) {
	return nil, nil
}

func (s *SmtpSender) send(sender string, recipient []string, body []byte) error {
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.addr, s.port),
		s.auth,
		sender,
		recipient,
		body,
	)
	return err
}

func (s *SmtpSender) SendWill(will *wl.Will) error {
	var recipients []string
	for _, recipient := range will.Recipients {
		recipients = append(recipients, recipient.Email)
	}
	body, err := createMailBody(will)
	if err != nil {
		return err
	}

	s.send("en4@nexo.cloud", recipients, body)

	return nil
}
