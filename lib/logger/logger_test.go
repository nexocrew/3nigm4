//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package logger

// Golang stdlib
import (
	"testing"
)

// External libs
import (
	"github.com/fatih/color"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger(color.New(color.FgGreen), "test", "MESSAGE", true, true)
	if l == nil {
		t.Fatalf("Unable to create a logger.\n")
	}

	coloredString := l.Color("Test this string\n")
	t.Log(coloredString)

	l.Println("This is a logging test.")
}

func TestNewLoggerFacility(t *testing.T) {
	lf := NewLogFacility("test", true, true)
	if lf == nil {
		t.Fatalf("Unable to create logging facility.\n")
	}
	lf.VerboseLog("Test verbose.\n")
	lf.MessageLog("Test message.\n")
	lf.WarningLog("Test warning.\n")
	lf.ErrorLog("Test error.\n")
	lf.CriticalLog("Test critical.\n")
}
