//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//
package logo

import (
	"fmt"
)

const (
	kANSIEscapedColorFormat = "\x1b[40m\x1b[%dm%s\x1b[0m" // defines the default format string to color a particular string.
	kANSIEscapedBoldFormat  = "\x1b[40m\x1b[1m%s\x1b[0m"
	kLogoSeparator          = `==============================`
	kLogoLine0              = ` ___  _  _  __  `
	kLogoLine1              = `(__ )( \( )/ ,) `
	kLogoLine2              = ` (_ \ )  ((_  _)`
	kLogoLine3              = `(___/(_)\_) (_) `
	kLogoLine4              = `3nigm4 project`
	kPrefixChar             = `*`
	kPostfixChar            = `*`
	kCopyright              = `GNU general public license`
	k3n4Notice              = `Secure communication tool`
)

const (
	kFrameColor      = 32
	kLogoTopColor    = 35
	kLogoBottomColor = 35
	kMainColor       = 35
	kTextColor       = 39
)

// Concatenate ANSI colors to a predefined
// string.
func colorString(str string, color int) string {
	return fmt.Sprintf(kANSIEscapedColorFormat, color, str)
}

// Use ANSI text properties to obtain bold
// characters.
func boldString(str string) string {
	return fmt.Sprintf(kANSIEscapedBoldFormat, str)
}

// Compose a line considering color and character type.
func composeLine(destination *string, str string, color int, bold bool) {
	body := colorString(str, color)
	if bold {
		body = boldString(body)
	}
	line := colorString(kPrefixChar, kFrameColor) + body + colorString(kPostfixChar, kFrameColor) + "\n"
	*destination += line
}

// Compose a separator to use in logo.
func composeSeparator(destination *string) {
	*destination += colorString(kLogoSeparator, kFrameColor) + "\n"
}

// Returns the logo string ready to be printed
// at application startup.
func Logo(componentName string, version string, startupInfos map[string]string) string {
	logo := "\n"
	composeSeparator(&logo)
	composeLine(&logo, kLogoLine0, kLogoTopColor, false)
	composeLine(&logo, kLogoLine1, kLogoTopColor, false)
	composeLine(&logo, kLogoLine2, kLogoTopColor, false)
	composeLine(&logo, kLogoLine3, kLogoTopColor, false)
	composeLine(&logo, kLogoLine3, kMainColor, true)
	composeLine(&logo, fmt.Sprintf("%7s%45.45s%5s", "", k3n4Notice, ""), kMainColor, false)
	composeSeparator(&logo)
	composeLine(&logo, fmt.Sprintf("%7s%45.45s%5s", "", componentName, ""), kTextColor, true)
	versionString := "Version: " + version
	composeLine(&logo, fmt.Sprintf("%7s%45.45s%5s", "", versionString, ""), kTextColor, true)
	composeLine(&logo, fmt.Sprintf("%7s%45.45s%5s", "", kCopyright, ""), kTextColor, false)
	for key := range startupInfos {
		infoString := key + ": " + startupInfos[key]
		composeLine(&logo, fmt.Sprintf("%7s%45.45s%5s", "", infoString, ""), kTextColor, false)
	}
	composeSeparator(&logo)

	return fmt.Sprintf("\x1b[0m%s\x1b[0m", logo)
}
