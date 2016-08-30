//
// 3nigm4 storageservice package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//
// This mock rpc servicec is used for tests purposes, should
// never be used in production environment. It's not
// concurrency safe and do not implement any performance
// optimisation logic.
//

package main

// Std golang libs
import (
	"encoding/hex"
	"fmt"
	"time"
)

// Internal libs
import (
	"github.com/nexocrew/3nigm4/lib/auth"
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

var (
	mockUserInfo = &auth.UserInfoResponseArg{
		Username: "userA",
		FullName: "User A",
		Email:    "usera@mail.com",
		Permissions: &auth.Permissions{
			SuperAdmin: false,
			Services: map[string]auth.Level{
				"storage": auth.LevelUser,
			},
		},
		LastSeen: time.Now(),
	}
	mockUserPassword = "passwordA"
)

type authMock struct {
	credentials map[string]string
	sessions    map[string]*auth.UserInfoResponseArg
}

func newAuthMock() (*authMock, error) {
	return &authMock{
		credentials: map[string]string{
			mockUserInfo.Username: mockUserPassword,
		},
		sessions: make(map[string]*auth.UserInfoResponseArg),
	}, nil
}

func (a *authMock) Login(username string, password string) ([]byte, error) {
	if password != a.credentials[username] {
		return nil, fmt.Errorf("wrong credentials")
	}
	token, err := ct.RandomBytesForLen(32)
	if err != nil {
		return nil, err
	}
	a.sessions[hex.EncodeToString(token)] = mockUserInfo
	return token, nil
}

func (a *authMock) Logout(token []byte) ([]byte, error) {
	delete(a.sessions, hex.EncodeToString(token))
	return token, nil
}

func (a *authMock) AuthoriseAndGetInfo(token []byte) (*auth.UserInfoResponseArg, error) {
	info, ok := a.sessions[hex.EncodeToString(token)]
	if !ok {
		return nil, fmt.Errorf("wrong session token")
	}
	return info, nil
}

func (a *authMock) Close() error {
	return nil
}
