//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

// Package logger manage a colored concurrent safe
// logger to substitute golang log package. This
// implementation is heavely base on github.com/fatih/color
// package (for all the coloring stuff).
package logger

import (
	"fmt"
	"sync"
	"time"
)

// Third party dependencies
import (
	"github.com/fatih/color"
)

// Logger single colored logger structure.
type Logger struct {
	mu        sync.Mutex   // ensures atomic writes; protects the following fields;
	color     *color.Color // actual color printer;
	prefix    string       // prefix string;
	tag       string       // logging type tag;
	timestamp bool         // activate timestamp;
	colored   bool         // colored print.
}

// NewLogger creates a new colored logger
// starting from a fatih color Color struct.
func NewLogger(color *color.Color,
	prefix, tag string,
	timestamp bool,
	colored bool) *Logger {
	return &Logger{
		prefix:    prefix,
		tag:       tag,
		timestamp: timestamp,
		color:     color,
		colored:   colored,
	}
}

// formatHeader create a rightly formatted
// header for logging messages.
func (l *Logger) formatHeader(logt string) string {
	var header string
	if l.prefix != "" {
		header += fmt.Sprintf("[%s] ", l.prefix)
	}
	if l.timestamp {
		header += fmt.Sprintf("[%s] ", time.Now().Format(time.UnixDate))
	}
	if l.prefix != "" ||
		l.timestamp {
		if logt != "" {
			c := l.color.SprintFunc()
			header += c(fmt.Sprintf("[%s]", logt)) + " "
		}
		c := color.New(color.BgBlack, color.FgHiWhite)
		funcColor := c.SprintFunc()

		return funcColor(header)
	}

	return ""
}

// Printf replace fmt.Printf() function.
func (l *Logger) Printf(f string, arg ...interface{}) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.colored {
		fmt.Print(l.formatHeader(l.tag))
		return l.color.Printf(f, arg...)
	}
	return fmt.Printf(f, arg...)
}

// Print replace fmt.Print() function.
func (l *Logger) Print(arg ...interface{}) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.colored {
		fmt.Print(l.formatHeader(l.tag))
		return l.color.Print(arg...)
	}
	return fmt.Print(arg...)
}

// Println prints a single line using specified
// color.
func (l *Logger) Println(arg ...interface{}) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.colored {
		fmt.Print(l.formatHeader(l.tag))
		return l.color.Println(arg...)
	}
	return fmt.Println(arg...)
}

// Sprintf colors a specific format string.
func (l *Logger) Sprintf(f string, arg ...interface{}) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.colored {
		funcColor := l.color.SprintfFunc()
		return funcColor(f, arg...)
	}
	return fmt.Sprintf(f, arg...)
}

// Color colors the argument string.
func (l *Logger) Color(str string) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.colored {
		funcColor := l.color.SprintFunc()
		return funcColor(str)
	}
	return str
}

// LogFacility logging facility
// structure.
type LogFacility struct {
	mu sync.Mutex // ensures atomic writes destined to different loggers;
	// log facility
	verbose  *Logger // verbose logging;
	message  *Logger // general message;
	warning  *Logger // warning message;
	errors   *Logger // error not critical;
	critical *Logger // critical errror;
	// additional infos
	prefix    string // prefix string;
	timestamp bool   // add timestamp string.
}

// NewLogFacility creates a new logging facility
// to manage differently colored logs. If a
// prefix is specified it'll be added to each
// logging message, a time stamp can be also
// printed on screen.
func NewLogFacility(prefix string, timestamp bool, colored bool) *LogFacility {
	return &LogFacility{
		prefix:    prefix,
		timestamp: timestamp,
		verbose: NewLogger(color.New(color.BgBlack, color.FgHiWhite),
			prefix,
			"VERBOSE",
			timestamp,
			colored),
		message: NewLogger(color.New(color.BgBlack, color.FgGreen),
			prefix,
			"MESSAGE",
			timestamp,
			colored),
		warning: NewLogger(color.New(color.BgBlack, color.FgYellow),
			prefix,
			"WARNING",
			timestamp,
			colored),
		errors: NewLogger(color.New(color.BgBlack, color.FgRed),
			prefix,
			"ERROR",
			timestamp,
			colored),
		critical: NewLogger(color.New(color.BgBlack, color.Bold, color.FgHiRed),
			prefix,
			"CRITICAL",
			timestamp,
			colored),
	}
}

// VerboseLog print white verbose logs.
func (f *LogFacility) VerboseLog(fmt string, arg ...interface{}) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.verbose.Printf(fmt, arg...)
}

// MessageLog generic green logs.
func (f *LogFacility) MessageLog(fmt string, arg ...interface{}) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.message.Printf(fmt, arg...)
}

// WarningLog prints warnings using yellow color.
func (f *LogFacility) WarningLog(fmt string, arg ...interface{}) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.warning.Printf(fmt, arg...)
}

// ErrorLog prints errors using red color.
func (f *LogFacility) ErrorLog(fmt string, arg ...interface{}) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.errors.Printf(fmt, arg...)
}

// CriticalLog prints critical logs using
// red color and bold font.
func (f *LogFacility) CriticalLog(fmt string, arg ...interface{}) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.critical.Printf(fmt, arg...)
}
