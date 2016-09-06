//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package authserver

// Golang std libs
import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
	"strconv"
	"time"
)

import (
	ty "github.com/nexocrew/3nigm4/lib/auth/types"
)

// Third party libs
import (
	"golang.org/x/crypto/bcrypt"
)

// Default bcrypt iterations
const kBCryptIterations = 10

// bcryptPassword helper function that creates a bcrypted password
// from a string. Usable to produce user's passwords in
// configuration tools. If all goes right returns a byte slice
// otherwise returns an error.
func bcryptPassword(pwd string) ([]byte, error) {
	pwdBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), kBCryptIterations)
	if err != nil {
		return nil, err
	}
	return pwdBytes, nil
}

// generateSessionToken generates a new unique token for the
// user session.
func generateSessionToken(username string) ([]byte, error) {
	// get time data
	now := time.Now()
	context := []byte(
		fmt.Sprintf("%s.%s.%s",
			username,
			strconv.FormatInt(now.Unix(), 10),
			strconv.FormatInt(now.UnixNano(), 10)),
	)
	// add 32 random bytes
	plain := make([]byte, len(context)+32)
	// copy time dependand string
	for idx := range context {
		plain[idx] = context[idx]
	}
	randBytes := plain[len(context):]
	if _, err := io.ReadFull(rand.Reader, randBytes); err != nil {
		return nil,
			fmt.Errorf("unable to get the right randomicity, unable to proceed: %s", err.Error())
	}
	// sha512-sum everything
	hash := sha512.Sum512(plain)
	return hash[:], nil
}

// Login the RPC required custom type.
type Login int

// Login RPC exposed functions it's create a session token
// after verifying that the username and password are already
// registered in the system.
func (t *Login) Login(args *ty.LoginRequestArg, response *ty.LoginResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Username == "" ||
		args.Password == "" {
		return fmt.Errorf("invalid username and password")
	}

	// query for user
	reference, err := client.GetUser(args.Username)
	if err != nil {
		return fmt.Errorf("unable to get %s user: %s", args.Username, err.Error())
	}
	if reference.IsDisabled == true {
		return fmt.Errorf("user is disabled, unable to proceed")
	}
	err = bcrypt.CompareHashAndPassword(reference.HashedPassword, []byte(args.Password))
	if err != nil {
		return fmt.Errorf("user not authenticated: %s", err.Error())
	}

	// create session token
	token, err := generateSessionToken(reference.Username)
	if err != nil {
		return err
	}
	// save to the database
	now := time.Now()
	err = client.SetSession(&ty.Session{
		Token:        token,
		Username:     reference.Username,
		LoginTime:    now,
		LastSeenTime: now,
		TimeToLive:   time.Duration(kTimeToLive) * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("unable to save session: %s", err.Error())
	}

	response.Token = token
	return nil
}

// Logout RPC exposed function logout a user, starting from a valid active
// session and remove all opened session related to that user.
func (t *Login) Logout(args *ty.LogoutRequestArg, response *ty.LogoutResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
	defer client.Close()

	if args.Token == nil {
		return fmt.Errorf("invalid session token")
	}

	// remove session, not verifying the actual validity, it
	// would not matter having to be removed if timeout has been
	// reached.
	err := client.RemoveSession(args.Token)
	if err != nil {
		return fmt.Errorf("unable to remove session: %s", err.Error())
	}

	response.Invalidated = args.Token
	return nil
}
