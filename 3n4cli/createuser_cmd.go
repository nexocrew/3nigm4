//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 12/12/2016
//

package main

// Golang std libs
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Internal dependencies
import (
	al "github.com/nexocrew/3nigm4/lib/auth"
)

// Third party libs
import (
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserCmd let's create a new user's docuement,
// this is a hidden function used to prepare db ready
// docs with correctly generated credentials.
var CreateUserCmd = &cobra.Command{
	Use:     "createuser",
	Short:   "Create a new user JSON coded record",
	Long:    "Create a correctly generated user document ready to be inserted in a database.",
	Example: "3n4cli createuser -u username",
	Hidden:  true,
}

func TrimLastChar(s string) string {
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

// createuser generate a new user's record starting from the provided
// username and asking for a usable password.
func createuser(cmd *cobra.Command, args []string) error {
	verbosePreRunInfos(cmd, args)

	reader := bufio.NewReader(os.Stdin)
	// get user data
	fmt.Printf("Insert username []: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = TrimLastChar(username)
	fmt.Printf("Insert user's full name []: ")
	fullname, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fullname = TrimLastChar(fullname)
	fmt.Printf("Insert user's email address []: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	email = TrimLastChar(email)

	// get user password
	fmt.Printf("Insert password []: ")
	pwd, err := gopass.GetPasswdMasked()
	if err != nil {
		return err
	}

	// get service label
	service := "all"
	fmt.Printf("Insert service label [all]: ")
	label, err := reader.ReadString('\n')
	if err == nil &&
		label != "" {
		service = TrimLastChar(label)
	}

	// get user permission
	fmt.Printf("Select user's permission level [superadmin, admin, user]: ")
	permission, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	var permStruct al.Permissions
	permStruct.Services = make(map[string]al.Level)
	permission = TrimLastChar(strings.ToLower(permission))
	switch {
	case permission == "superadmin":
		permStruct.SuperAdmin = true
	case permission == "admin":
		permStruct.SuperAdmin = false
		permStruct.Services[service] = al.LevelAdmin
	case permission == "user":
		permStruct.SuperAdmin = false
		permStruct.Services[service] = al.LevelUser
	default:
		return fmt.Errorf("unknown perrmission level \"%s\", unable to proceed", permission)
	}

	hexedPwd := hexComposedPassword(username, pwd)
	bcryptedPwd, err := bcrypt.GenerateFromPassword([]byte(hexedPwd), 10)
	if err != nil {
		return err
	}

	user := &al.User{
		Username:       username,
		FullName:       fullname,
		Email:          email,
		Permissions:    permStruct,
		HashedPassword: bcryptedPwd,
	}
	encoded, err := json.Marshal(user)
	if err != nil {
		return err
	}
	fmt.Printf("Prepared user:\n%s\n", string(encoded))

	return nil
}
