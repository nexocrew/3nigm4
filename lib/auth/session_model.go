//
// 3nigm4 auth package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package auth

// Golang std libs
import (
	"encoding/hex"
	"fmt"
	"time"
)

const (
	kTimeToLive = 15 // minutes to live for a session between accesses.
)

// SessionAuth RPC required custom type (using int arbitrarely).
type SessionAuth int

// VoidResponseArg empty return value.
type VoidResponseArg struct{}

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
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
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
	Username    string       // the session related username;
	FullName    string       // the user full name;
	Email       string       // the user email address;
	Permissions *Permissions // user associated permissions;
	LastSeen    time.Time    // last seen info.
}

// UserInfo RPC exposed function verify a session token
// and returns the user associated data (from the User struct).
// Notice that this function will update the "last seen" time
// stamp as the Authenticate do.
func (s *SessionAuth) UserInfo(args *AuthenticateRequestArg, response *UserInfoResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
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
	response.Permissions = &user.Permissions
	response.LastSeen = userResponse.LastSeenTime

	return nil
}

//
// Superadmin behaviour: the following functions are intended to
// implement administrative tasks like creating or removing users,
// update user's permissions or logout all users.
//

// UpserUserRequestArg request to upsert user data.
type UpserUserRequestArg struct {
	Token []byte // the authentication token;
	User  User   // the user record to be updated.
}

// UpsertUser is an RPC exposed function used to add or update a user in
// the authentication database. If the user is not already present it'll
// be added, otherwise it will be updated. Only Super-Admins will be able
// to use this function.
func (s *SessionAuth) UpsertUser(args *UpserUserRequestArg, response *VoidResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Token == nil {
		return fmt.Errorf("invalid nil token data")
	}

	userinfo := UserInfoResponseArg{}
	err := s.UserInfo(&AuthenticateRequestArg{
		Token: args.Token,
	}, &userinfo)
	if err != nil {
		return err
	}
	// check for superadmin
	if userinfo.Permissions.SuperAdmin != true {
		return fmt.Errorf("user not authorised to perform this operation, contact system admin")
	}
	// do actual job
	err = client.SetUser(&args.User)
	if err != nil {
		return fmt.Errorf("unable to upsert user %s: %s", args.User.Username, err.Error())
	}

	return nil
}

// RemoveUserRequestArg request for remove an existing
// user.
type RemoveUserRequestArg struct {
	Token    []byte // the authentication token;
	Username string // the user to be removed.
}

// RemoveUser is an RPC exposed function that removes an existing user
// from the authentication db.
func (s *SessionAuth) RemoveUser(args *RemoveUserRequestArg, response *VoidResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Token == nil {
		return fmt.Errorf("invalid nil token data")
	}
	// check for username
	if args.Username == "" {
		return fmt.Errorf("invalid username: unable to process requesto for nil username")
	}
	// get user infos
	userinfo := UserInfoResponseArg{}
	err := s.UserInfo(&AuthenticateRequestArg{
		Token: args.Token,
	}, &userinfo)
	if err != nil {
		return err
	}
	// check for superadmin
	if userinfo.Permissions.SuperAdmin != true {
		return fmt.Errorf("user not authorised to perform this operation, contact system admin")
	}
	// do actual job
	err = client.RemoveUser(args.Username)
	if err != nil {
		return fmt.Errorf("unable to remove user: %s", err.Error())
	}

	return nil
}

// KickOutAllSessions is an RPC exposed function that remove all active sessions from
// the authentication database.
func (s *SessionAuth) KickOutAllSessions(args *AuthenticateRequestArg, response *VoidResponseArg) error {
	// check for session
	if dbclient == nil {
		return fmt.Errorf("invalid db session, unable to proceed")
	}
	client := dbclient.Copy()
	defer client.Close()

	// check for arguments
	if args == nil ||
		args.Token == nil {
		return fmt.Errorf("invalid nil token data")
	}
	// get user infos
	userinfo := UserInfoResponseArg{}
	err := s.UserInfo(&AuthenticateRequestArg{
		Token: args.Token,
	}, &userinfo)
	if err != nil {
		return err
	}
	// check for superadmin
	if userinfo.Permissions.SuperAdmin != true {
		return fmt.Errorf("user not authorised to perform this operation, contact system admin")
	}
	// do actual job
	err = client.RemoveAllSessions()
	if err != nil {
		return fmt.Errorf("unable to remove all sessions: %s", err.Error())
	}

	return nil
}
