//
// 3nigm4 sendermock package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

// Package sendermock implements a mock struct to
// send emails in unit-tests.
package sendermock

// Golang std pkgs
import (
	"sync"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

type MockSender struct {
	Sended map[string]SendedMessage
	mtx    sync.Mutex
}

type SendedMessage struct {
	Email          *ct.Email
	Subject        string
	FromAddress    string
	AttachmentName string
}

func NewMockSender() *MockSender {
	return &MockSender{
		Sended: make(map[string]SendedMessage),
	}
}

func (s *MockSender) SendEmail(content *ct.Email, fromAddress, subject, attachmentName string) error {
	s.mtx.Lock()
	s.Sended[content.Recipient] = SendedMessage{
		Email:          content,
		Subject:        subject,
		FromAddress:    fromAddress,
		AttachmentName: attachmentName,
	}
	s.mtx.Unlock()
	return nil
}
