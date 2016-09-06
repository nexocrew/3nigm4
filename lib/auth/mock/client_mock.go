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

package authmock

// Std golang libs
import (
	"encoding/hex"
	"fmt"
	"time"
)

// Internal libs
import (
	ty "github.com/nexocrew/3nigm4/lib/auth/types"
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

var (
	MockUserInfo = &ty.UserInfoResponseArg{
		Username: "userA",
		FullName: "User A",
		Email:    "usera@mail.com",
		Permissions: &ty.Permissions{
			SuperAdmin: false,
			Services: map[string]ty.Level{
				"storage": ty.LevelUser,
			},
		},
		LastSeen: time.Now(),
	}
	MockUserPassword = "passwordA"
)

type AuthMock struct {
	credentials map[string]string
	sessions    map[string]*ty.UserInfoResponseArg
}

func NewAuthMock() (*AuthMock, error) {
	return &AuthMock{
		credentials: map[string]string{
			MockUserInfo.Username: MockUserPassword,
		},
		sessions: make(map[string]*ty.UserInfoResponseArg),
	}, nil
}

func (a *AuthMock) Login(username string, password string) ([]byte, error) {
	if password != a.credentials[username] {
		return nil, fmt.Errorf("wrong credentials")
	}
	token, err := ct.RandomBytesForLen(32)
	if err != nil {
		return nil, err
	}
	a.sessions[hex.EncodeToString(token)] = MockUserInfo
	return token, nil
}

func (a *AuthMock) Logout(token []byte) ([]byte, error) {
	delete(a.sessions, hex.EncodeToString(token))
	return token, nil
}

func (a *AuthMock) AuthoriseAndGetInfo(token []byte) (*ty.UserInfoResponseArg, error) {
	info, ok := a.sessions[hex.EncodeToString(token)]
	if !ok {
		return nil, fmt.Errorf("wrong session token")
	}
	return info, nil
}

func (a *AuthMock) Close() error {
	return nil
}
