//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"encoding/hex"
	"fmt"
	"time"
)

const (
	kTimeToLive = 15 // minutes to live for a session between accesses.
)

// Token the RPC required custom type.
type SessionAuth int

// AuthenticateRequestArg define the RPC request struct
type AuthenticateRequestArg struct {
	Token []byte // the authentication token.
}

// AuthenticateResponseArg the returned auth structure.
type AuthenticateResponseArg struct {
	Username     string    // the session related username;
	LastSeenTime time.Time // last connection from the user.
}

// sessionTimeValid verify the time range between last seen
// time and now, if it exceed the session expiration time (15 min)
// it returns true otherwise false.
func sessionTimeValid(now, lastSeen *time.Time, timeToLive time.Duration) bool {
	if now.Sub(*lastSeen) > timeToLive {
		return false
	}
	return true
}

// Authenticate RPC exposed functions verify a session token
// and returns the userid to authenticate user required
// operations.
func (s *SessionAuth) Authenticate(args *AuthenticateRequestArg, response *AuthenticateResponseArg) error {
	// check for session
	if arguments.dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := arguments.dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Token == nil {
		return fmt.Errorf("invalid nil token data")
	}

	// find session in the db
	session, err := client.GetSession(args.Token)
	if err != nil {
		return fmt.Errorf("unable to get required session %s: %s", hex.EncodeToString(args.Token), err.Error())
	}

	// validate time vars
	now := time.Now()
	if sessionTimeValid(&now, &session.LastSeenTime, session.TimeToLive) == false {
		client.RemoveSession(args.Token)
		return fmt.Errorf("session is expired")
	}

	response.Username = session.Username
	response.LastSeenTime = session.LastSeenTime

	// update last seen
	session.LastSeenTime = now
	err = client.SetSession(session)
	if err != nil {
		return fmt.Errorf("unable to update session last seen time stamp: %s", err.Error())
	}

	return nil
}

// UserInfoResponseArg the returned authenticated  user
// data.
type UserInfoResponseArg struct {
	Username string    // the session related username;
	FullName string    // the user full name;
	Email    string    // the user email address;
	LastSeen time.Time // last seen info.
}

// UserInfo RPC exposed function verify a session token
// and returns the user associated data (from the User struct).
// Notice that this function will update the "last seen" time
// stamp as the Authenticate do.
func (s *SessionAuth) UserInfo(args *AuthenticateRequestArg, response *UserInfoResponseArg) error {
	// check for session
	if arguments.dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := arguments.dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Token == nil {
		return fmt.Errorf("invalid nil token data")
	}

	userResponse := AuthenticateResponseArg{}
	err := s.Authenticate(args, &userResponse)
	if err != nil {
		return err
	}

	// get user info
	user, err := client.GetUser(userResponse.Username)
	if err != nil {
		return fmt.Errorf("unable to get required user %s: %s", userResponse.Username, err.Error())
	}

	response.Username = user.Username
	response.Email = user.Email
	response.FullName = user.FullName
	response.LastSeen = userResponse.LastSeenTime

	return nil
}
