//
// 3nigm4 ishtmtypes package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

// Package ishtmtypes defines common types and utility
// functions for the ishtm package.
package ishtmtypes

// Golang std pkgs
import (
	"fmt"
)

// ComposeDbAddress compose a string starting from dbArgs slice.
func ComposeDbAddress(args *DbArgs) string {
	dbAccess := fmt.Sprintf("mongodb://%s:%s@", args.User, args.Password)
	for idx, addr := range args.Addresses {
		dbAccess += addr
		if idx != len(args.Addresses)-1 {
			dbAccess += ","
		}
	}
	dbAccess += fmt.Sprintf("/?authSource=%s", args.AuthDb)
	return dbAccess
}
