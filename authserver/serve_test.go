//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
package main

// Std Golang libs
import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"testing"
	"time"
)

// Internal dependencies.
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	"github.com/nexocrew/3nigm4/lib/itm"
	"github.com/nexocrew/3nigm4/lib/logger"
	wq "github.com/nexocrew/3nigm4/lib/workingqueue"
)

func mockStartup(arguments *args) (auth.Database, error) {
	mockdb := newMockDb(&auth.DbArgs{
		Addresses: strings.Split(arguments.dbAddresses, ","),
		User:      arguments.dbUsername,
		Password:  arguments.dbPassword,
		AuthDb:    arguments.dbAuth,
	})

	log.MessageLog("Mockdb %s successfully connected.\n", arguments.dbAddresses)

	err := prepareMockDb(mockdb)
	if err != nil {
		return nil, err
	}

	return mockdb, nil
}

// prepareMockDb define all credentials used in tests.
func prepareMockDb(db *mockdb) error {
	// add test user
	hash, err := bcryptPassword("passwordA")
	if err != nil {
		return err
	}
	err = db.SetUser(&auth.User{
		Username:       "userA",
		FullName:       "user A",
		Email:          "userA@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: auth.Permissions{
			SuperAdmin: false,
			Services: map[string]auth.Level{
				"test": auth.LevelAdmin,
			},
		},
	})
	if err != nil {
		return err
	}

	// superadmin
	hash, err = bcryptPassword("passwordS")
	if err != nil {
		return err
	}
	err = db.SetUser(&auth.User{
		Username:       "asuperadmin",
		FullName:       "Super Admin",
		Email:          "superadmin@email.com",
		IsDisabled:     false,
		HashedPassword: hash,
		Permissions: auth.Permissions{
			SuperAdmin: true,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	// start up logging facility
	log = logger.NewLogFacility("authserver", true, true)

	arguments = args{
		verbose:     true,
		colored:     true,
		dbAddresses: fmt.Sprintf("%s:%d", itm.S().DbAddress(), itm.S().DbPort()),
		dbUsername:  itm.S().DbUserName(),
		dbPassword:  itm.S().DbPassword(),
		dbAuth:      itm.S().DbAuth(),
		address:     "127.0.0.1",
		port:        17743,
	}
	databaseStartup = mockStartup

	var errorCounter wq.AtomicCounter
	errorChan := make(chan error, 0)
	var lastError error
	go func() {
		for {
			select {
			case err, _ := <-errorChan:
				errorCounter.Add(1)
				lastError = err
			}
		}
	}()
	// startup service
	go func(ec chan error) {
		err := serve(ServeCmd, nil)
		if err != nil {
			ec <- err
			return
		}
	}(errorChan)
	// the following timeout time is used to ensure
	// that all goroutines have compleated their
	// processing life (especially to verify that
	// no error is returned by concurrent server
	// startup). 3 seconds is an arbitrary, experimentally
	// defined, time on some slower systems it can be not
	// enought.
	ticker := time.Tick(3 * time.Second)
	timeoutCounter := wq.AtomicCounter{}
	go func() {
		for {
			select {
			case <-ticker:
				timeoutCounter.Add(1)
			}
		}
	}()
	// infinite loop:
	for {
		if timeoutCounter.Value() != 0 {
			break
		}
		if errorCounter.Value() != 0 {
			log.ErrorLog("Error returned: %s.\n", lastError)
			os.Exit(1)
		}
		time.Sleep(50 * time.Millisecond)
	}

	os.Exit(m.Run())
}

func TestRPCServe(t *testing.T) {
	// test RPC calls
	address := fmt.Sprintf("%s:%d", arguments.address, arguments.port)
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		t.Fatalf("Unable to connect to RPC server: %s.\n", err.Error())
	}
	defer client.Close()
	t.Logf("Client connected to %s.\n", address)

	// login
	var loginResponse auth.LoginResponseArg
	err = client.Call("Login.Login", &auth.LoginRequestArg{
		Username: "userA",
		Password: "passwordA",
	}, &loginResponse)
	if err != nil {
		t.Fatalf("Unable to perform login: %s.\n", err.Error())
	}
	if loginResponse.Token == nil {
		t.Fatalf("Invalid returned token, should not be nil.\n")
	}
	if len(loginResponse.Token) != 64 {
		t.Fatalf("Invalid token size, having %d expecting 64.\n", len(loginResponse.Token))
	}
	// session validation
	var sessionResponse auth.AuthenticateResponseArg
	err = client.Call("SessionAuth.Authenticate", &auth.AuthenticateRequestArg{
		Token: loginResponse.Token,
	}, &sessionResponse)
	if err != nil {
		t.Fatalf("Unable to authenticate with token: %s.\n", err.Error())
	}
	if sessionResponse.Username != "userA" {
		t.Fatalf("Wrong username, having %s expecting \"userA\".\n", sessionResponse.Username)
	}
	// userinfo
	var userinfoResponse auth.UserInfoResponseArg
	err = client.Call("SessionAuth.UserInfo", &auth.AuthenticateRequestArg{
		Token: loginResponse.Token,
	}, &userinfoResponse)
	if err != nil {
		t.Fatalf("Unable to get userinfo: %s.\n", err.Error())
	}
	if userinfoResponse.Permissions.Services["test"] != auth.LevelAdmin {
		t.Fatalf("Wrong auth level, having %d expecting %d.\n", userinfoResponse.Permissions.Services["test"], auth.LevelAdmin)
	}
	// logout
	var logoutResponse auth.LogoutResponseArg
	err = client.Call("Login.Logout", &auth.LogoutRequestArg{
		Token: loginResponse.Token,
	}, &logoutResponse)
	if err != nil {
		t.Fatalf("Unable to logout the user: %s.\n", err.Error())
	}
	if bytes.Compare(logoutResponse.Invalidated, loginResponse.Token) != 0 {
		t.Fatalf("Unexpected token, having %s, expecting %s.\n", hex.EncodeToString(logoutResponse.Invalidated), hex.EncodeToString(loginResponse.Token))
	}
}

func TestRPCServeSuperAdmin(t *testing.T) {
	// test RPC calls
	address := fmt.Sprintf("%s:%d", arguments.address, arguments.port)
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		t.Fatalf("Unable to connect to RPC server: %s.\n", err.Error())
	}
	defer client.Close()
	t.Logf("Client connected to %s.\n", address)

	// login
	var loginResponse auth.LoginResponseArg
	err = client.Call("Login.Login", &auth.LoginRequestArg{
		Username: "asuperadmin",
		Password: "passwordS",
	}, &loginResponse)
	if err != nil {
		t.Fatalf("Unable to perform login: %s.\n", err.Error())
	}
	if loginResponse.Token == nil {
		t.Fatalf("Invalid returned token, should not be nil.\n")
	}
	if len(loginResponse.Token) != 64 {
		t.Fatalf("Invalid token size, having %d expecting 64.\n", len(loginResponse.Token))
	}
	// create user
	var voidResponse auth.VoidResponseArg
	hash, err := bcryptPassword("passwordB")
	if err != nil {
		t.Fatalf("Unable to bcrypt password: %s.\n", err.Error())
	}
	err = client.Call("SessionAuth.UpsertUser", &auth.UpserUserRequestArg{
		Token: loginResponse.Token,
		User: auth.User{
			Username:       "userB",
			FullName:       "user B",
			Email:          "userB@email.com",
			IsDisabled:     false,
			HashedPassword: hash,
			Permissions: auth.Permissions{
				SuperAdmin: false,
				Services: map[string]auth.Level{
					"test": auth.LevelAdmin,
				},
			},
		},
	}, &voidResponse)
	if err != nil {
		t.Fatalf("Unable to upsert user: %S.\n", err.Error())
	}
	// remove user
	err = client.Call("SessionAuth.RemoveUser", &auth.RemoveUserRequestArg{
		Token:    loginResponse.Token,
		Username: "userB",
	}, &voidResponse)
	if err != nil {
		t.Fatalf("Unable to remove user: %s.\n", err.Error())
	}
	// kick off all sessions
	err = client.Call("SessionAuth.KickOutAllSessions", &auth.AuthenticateRequestArg{
		Token: loginResponse.Token,
	}, &voidResponse)
	if err != nil {
		t.Fatalf("Unable to remove all sessions: %s.\n", err.Error())
	}

	// logout
	var logoutResponse auth.LogoutResponseArg
	err = client.Call("Login.Logout", &auth.LogoutRequestArg{
		Token: loginResponse.Token,
	}, &logoutResponse)
	if err == nil {
		t.Fatalf("After resetting all session also super admin session should be gone.\n")
	}
}
