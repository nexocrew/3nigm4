//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"testing"
)

// Third party libs
import (
	"github.com/spf13/cobra"
)

// No perticular tests are implemented for this
// package as it uses all functions exposed by
// heavely tested packages. This exec fo not
// have any data manipulating implementation all
// extenal function actually implement the used
// logic.
// For all other tests refer to the specific packages.

func TestViperLabel(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Some short description",
		Long:  "Some very long and long description.",
	}
	secondaryCmd := &cobra.Command{
		Use:   "secondary",
		Short: "Some short description",
		Long:  "Some very long and long description.",
	}
	thirdCmd := &cobra.Command{
		Use:   "third",
		Short: "Some short description",
		Long:  "Some very long and long description.",
	}
	fourthCmd := &cobra.Command{
		Use:   "fourth",
		Short: "Some short description",
		Long:  "Some very long and long description.",
	}
	rootCmd.AddCommand(secondaryCmd)
	secondaryCmd.AddCommand(thirdCmd)
	thirdCmd.AddCommand(fourthCmd)

	label := viperLabel(fourthCmd, "verbose")
	reference := "secondary.third.fourth.verbose"
	if label != reference {
		t.Fatalf("Unexpected result, having %s expecting %s.\n", label, reference)
	}

	label = viperLabel(secondaryCmd, "verbose")
	reference = "secondary.verbose"
	if label != reference {
		t.Fatalf("Unexpected result, having %s expecting %s.\n", label, reference)
	}
}

func TestViperLabelSingleCmd(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Some short description",
		Long:  "Some very long and long description.",
	}

	label := viperLabel(rootCmd, "verbose")
	reference := "verbose"
	if label != reference {
		t.Fatalf("Unexpected result, having %s expecting %s.\n", label, reference)
	}
}
