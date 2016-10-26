//
// 3nigm4 sender package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

// Package sender implements a generic interface for
// sending email messages.
package sender

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

// Sender interface represent sending objects.
type Sender interface {
	SendEmail(*ct.Email, string, string, string) error // function to actually send email messages.
}
