//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package auth

// Golang std libs
import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSessionAuth(t *testing.T) {
	// startup mock and global vars
	dbclient = newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: false,
			Services: map[string]Level{
				"test": LevelAdmin,
			},
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	var l Login
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	// session auth
	var s SessionAuth
	sessionValidation := AuthenticateResponseArg{}
	now := time.Now()
	err = s.Authenticate(&AuthenticateRequestArg{
		Token: response.Token,
	}, &sessionValidation)
	if err != nil {
		t.Fatalf("Unable to validate session %s: %s.\n", hex.EncodeToString(response.Token), err.Error())
	}
	if sessionValidation.Username != "userA" {
		t.Fatalf("Invalid username: having %s expecting \"userA\".\n", sessionValidation.Username)
	}
	if (sessionValidation.LastSeenTime.Hour()-now.Hour() != 0 &&
		sessionValidation.LastSeenTime.Hour()-now.Hour() != 1) ||
		(sessionValidation.LastSeenTime.Minute()-now.Minute() != 0 &&
			sessionValidation.LastSeenTime.Minute()-now.Minute() != 1) {
		t.Fatalf("Unexpected time should be similar to now %s but found %s.\n", now.String(), sessionValidation.LastSeenTime.String())
	}
}

func TestSessionAuthExpired(t *testing.T) {
	// startup mock and global vars
	mock := newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})
	dbclient = mock

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: false,
			Services: map[string]Level{
				"test": LevelAdmin,
			},
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	var l Login
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	now := time.Now()
	_, ok := mock.sessionStorage[hex.EncodeToString(response.Token)]
	if !ok {
		t.Fatalf("Unable to find created session %s.\n", hex.EncodeToString(response.Token))
	}
	mock.sessionStorage[hex.EncodeToString(response.Token)].LoginTime = time.Unix(now.Unix()-(16*60), 0)
	mock.sessionStorage[hex.EncodeToString(response.Token)].LastSeenTime = time.Unix(now.Unix()-(16*60), 0)

	// session auth
	var s SessionAuth
	sessionValidation := AuthenticateResponseArg{}
	err = s.Authenticate(&AuthenticateRequestArg{
		Token: response.Token,
	}, &sessionValidation)
	if err == nil {
		t.Fatalf("Session must be expired.\n")
	}
}

func TestSessionGetUserInfo(t *testing.T) {
	// startup mock and global vars
	dbclient = newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: true,
			Services: map[string]Level{
				"test": LevelAdmin,
			},
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	var l Login
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	// session auth
	var s SessionAuth
	sessionValidation := AuthenticateResponseArg{}
	err = s.Authenticate(&AuthenticateRequestArg{
		Token: response.Token,
	}, &sessionValidation)
	if err != nil {
		t.Fatalf("Unable to validate session %s: %s.\n", hex.EncodeToString(response.Token), err.Error())
	}

	// userinfo auth
	userinfo := UserInfoResponseArg{}
	err = s.UserInfo(&AuthenticateRequestArg{
		Token: response.Token,
	}, &userinfo)
	if err != nil {
		t.Fatalf("Unable to get userinfo from session %s: %s.\n", hex.EncodeToString(response.Token), err.Error())
	}
	if userinfo.Username != "userA" {
		t.Fatalf("Invalid username: having %s expecting \"userA\".\n", userinfo.Username)
	}
	if userinfo.FullName != "user A" {
		t.Fatalf("Invalid full name having %s expecting \"user A\".\n", userinfo.FullName)
	}
	if userinfo.Email != "userA@email.com" {
		t.Fatalf("Invalid email address having %s expecting \"userA@email.com\".\n", userinfo.Email)
	}
	if userinfo.LastSeen.Equal(sessionValidation.LastSeenTime) == true {
		t.Fatalf("Last seen should be updated by the token validation func but found equal values.\n")
	}

	// check for permissions
	if userinfo.Permissions.SuperAdmin != true {
		t.Fatalf("Unexpected not granted superadmin permission.\n")
	}
	level, ok := userinfo.Permissions.Services["test"]
	if !ok {
		t.Fatalf("Unable to find right permission for \"test\" service.\n")
	}
	if level != LevelAdmin {
		t.Fatalf("Unexpected level for \"test\" service: having %d expecting %d.\n", level, LevelAdmin)
	}
}

func TestSessionAddUser(t *testing.T) {
	// startup mock and global vars
	dbclient = newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: true,
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	var l Login
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	// add user
	var s SessionAuth
	voidResonse := VoidResponseArg{}
	err = s.UpsertUser(&UpserUserRequestArg{
		Token: response.Token,
		User: User{
			Username:       "createdusr",
			FullName:       "Newly created user",
			Email:          "createdusr@email.com",
			IsDisabled:     false,
			HashedPassword: hash,
			Permissions: Permissions{
				SuperAdmin: false,
				Services: map[string]Level{
					"test": LevelUser,
				},
			},
		},
	}, &voidResonse)
	if err != nil {
		t.Fatalf("Unable to add user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("createdusr")

	user, err := dbclient.GetUser("createdusr")
	if err != nil {
		t.Fatalf("User \"createdusr\" not retrived from database: %s.\n", err.Error())
	}
	if user.FullName != "Newly created user" {
		t.Fatalf("Unexpected full name: having %s expecting \"Newly created user\".\n", user.FullName)
	}
}

func TestSessionRemoveUser(t *testing.T) {
	// startup mock and global vars
	dbclient = newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: true,
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	var l Login
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	// add user
	var s SessionAuth
	voidResonse := VoidResponseArg{}
	err = s.UpsertUser(&UpserUserRequestArg{
		Token: response.Token,
		User: User{
			Username:       "createdusr",
			FullName:       "Newly created user",
			Email:          "createdusr@email.com",
			IsDisabled:     false,
			HashedPassword: hash,
			Permissions: Permissions{
				SuperAdmin: false,
				Services: map[string]Level{
					"test": LevelUser,
				},
			},
		},
	}, &voidResonse)
	if err != nil {
		t.Fatalf("Unable to add user: %s.\n", err.Error())
	}

	// remove it
	err = s.RemoveUser(&RemoveUserRequestArg{
		Token:    response.Token,
		Username: "createdusr",
	}, &voidResonse)
	if err != nil {
		t.Fatalf("Unable to remove user: %s.\n", err.Error())
	}
	_, err = dbclient.GetUser("createdusr")
	if err == nil {
		t.Fatalf("User must not exist any more after deletion.\n")
	}
}

func TestSessionKickOutSessions(t *testing.T) {
	// startup mock and global vars
	mock := newMockDb(&DbArgs{
		Addresses: strings.Split("127.0.0.1:27017,192.168.0.1:27017", ","),
		User:      "username",
		Password:  "password",
		AuthDb:    "admin",
	})
	dbclient = mock

	var l Login
	var s SessionAuth
	for idx := 0; idx < 10; idx++ {
		// add test user
		pwd := fmt.Sprintf("password%d", idx)
		hash, err := bcryptPassword(pwd)
		if err != nil {
			t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
		}
		username := fmt.Sprintf("user%d", idx)
		err = dbclient.SetUser(&User{
			Username:       username,
			FullName:       fmt.Sprintf("user %d", idx),
			Email:          "user@email.com",
			IsDisabled:     false,
			HashedPassword: hash,
			Permissions: Permissions{
				SuperAdmin: false,
				Services: map[string]Level{
					"test": LevelUser,
				},
			},
		})
		if err != nil {
			t.Fatalf("Unable to set user: %s.\n", err.Error())
		}
		defer dbclient.RemoveUser(username)

		// login func
		response := &LoginResponseArg{}
		err = l.Login(&LoginRequestArg{
			Username: username,
			Password: pwd,
		}, response)
		if err != nil {
			t.Fatalf("Unable to login user: %s.\n", err.Error())
		}
	}

	// check sessions
	if len(mock.sessionStorage) != 10 {
		t.Fatalf("Unexpected number of sessions, having %d expecting %d.\n", len(mock.sessionStorage), 10)
	}

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: Permissions{
			SuperAdmin: true,
		},
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer dbclient.RemoveUser("userA")

	// login func
	response := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, response)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}

	void := &VoidResponseArg{}
	err = s.KickOutAllSessions(&AuthenticateRequestArg{
		Token: response.Token,
	}, void)
	if err != nil {
		t.Fatalf("Unexpected error while removing all sessions: %s.\n", err.Error())
	}

	// check sessions again
	if len(mock.sessionStorage) != 0 {
		t.Fatalf("Unexpected number of sessions, having %d expecting %d.\n", len(mock.sessionStorage), 0)
	}
}
