//
// 3nigm4 mail package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package mail

type MailManager interface {
	SendMail(string, string, []byte) error
}
