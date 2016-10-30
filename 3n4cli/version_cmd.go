//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Golang std libs
import (
	"fmt"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/logger"
	"github.com/nexocrew/3nigm4/lib/logo"
	ver "github.com/nexocrew/3nigm4/lib/version"
)

// Third party libs
import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// VersionCmd shows client version and 3n4 logo.
var VersionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show client version",
	Long:    "Show client version and 3n4 logo.",
	Example: "3n4cli version",
	Run:     version,
}

func printLogo() {
	// print logo
	fmt.Printf("%s", logo.Logo("3n4 cli client", ver.V().VersionString(), nil))
}

// version show client version and 3n4 logo.
func version(cmd *cobra.Command, args []string) {
	printLogo()
	fmt.Printf("Command line client to access 3nigm4 services.\n")
	lg := logger.NewLogger(color.New(color.BgBlack, color.FgYellow, color.Underline), "", "", false, true)
	fmt.Printf("Take a look at %s for project details and docs.", lg.Color("www.3n4.io"))
}
