//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestSessionAuth(t *testing.T) {
	// startup mock and global vars
	arguments = args{
		dbAddresses: "127.0.0.1:27017,192.168.0.1:27017",
		dbUsername:  "username",
		dbPassword:  "password",
		address:     "0.0.0.0",
		port:        7300,
	}
	arguments.dbclient = newMockDb(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = arguments.dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer arguments.dbclient.RemoveUser("userA")

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
	arguments = args{
		dbAddresses: "127.0.0.1:27017,192.168.0.1:27017",
		dbUsername:  "username",
		dbPassword:  "password",
		address:     "0.0.0.0",
		port:        7300,
	}
	mock := newMockDb(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})
	arguments.dbclient = mock

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = arguments.dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer arguments.dbclient.RemoveUser("userA")

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
	arguments = args{
		dbAddresses: "127.0.0.1:27017,192.168.0.1:27017",
		dbUsername:  "username",
		dbPassword:  "password",
		address:     "0.0.0.0",
		port:        7300,
	}
	arguments.dbclient = newMockDb(&dbArgs{
		addresses: strings.Split(arguments.dbAddresses, ","),
		user:      arguments.dbUsername,
		password:  arguments.dbPassword,
		authDb:    arguments.dbAuth,
	})

	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		t.Fatalf("Unable to produce bcrypted password: %s.\n", err.Error())
	}
	err = arguments.dbclient.SetUser(&User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
	})
	if err != nil {
		t.Fatalf("Unable to set user: %s.\n", err.Error())
	}
	defer arguments.dbclient.RemoveUser("userA")

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
}
