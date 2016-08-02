//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Golang std libs
import (
	"strings"
	"testing"
)

func TestLoginRegularUser(t *testing.T) {
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
	if response.Token == nil ||
		len(response.Token) == 0 {
		t.Fatalf("Unexpected token, should be not nil.\n")
	}
	defer arguments.dbclient.RemoveSession(response.Token)
	t.Logf("Token: %v.\n", response.Token)
}

func TestLoginIvalidUser(t *testing.T) {
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
		Username: "userB",
		Password: "passwordA",
	}, response)
	if err == nil {
		t.Fatalf("Unknown users should not login.\n")
	}
	if response.Token != nil {
		t.Fatalf("Token response should be nil.\n")
	}
}

func TestLoginDisabledUser(t *testing.T) {
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
		IsDisabled:     true,
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
	if err == nil {
		t.Fatalf("Disabled users should not login.\n")
	}
	if response.Token != nil {
		t.Fatalf("Token response should be nil.\n")
	}
}

func TestLoginAndLogoutOnRegularUser(t *testing.T) {
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
	loginResponse := &LoginResponseArg{}
	err = l.Login(&LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, loginResponse)
	if err != nil {
		t.Fatalf("Unable to login user: %s.\n", err.Error())
	}
	if loginResponse.Token == nil ||
		len(loginResponse.Token) == 0 {
		t.Fatalf("Unexpected token, should be not nil.\n")
	}

	// logout
	logoutResponse := &LogoutResponse{}
	err = l.Logout(&LogoutRequest{
		Token: loginResponse.Token,
	}, logoutResponse)
	if err != nil {
		t.Fatalf("Unable to logout user.\n")
	}

	if _, err = arguments.dbclient.GetSession(loginResponse.Token); err == nil {
		t.Fatalf("Session should not be present but is still there.\n")
	}
}
