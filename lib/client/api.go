//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//

// Package client implements all client used functions
// that relates to the interaction with external services
// typically APIs.
package client

// Std golib dependencies
import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Internal dependencies
import (
	"github.com/nexocrew/3nigm4/lib/messages"
)

// Client represent the base in memory structure
// used to represent the user acting and all service
// related infos.
type Client struct {
	Sessions        []messages.SessionKeys        `json:"-" xml:"-"` // client opened sessions;
	PrivateKey      *openpgp.Entity               `json:"-" xml:"-"` // user's in memory private key;
	PublicKey       *openpgp.Entity               `json:"-" xml:"-"` // user's in memory public key;
	RecipientsCache map[string]openpgp.EntityList `json:"-" xml:"-"` // cached recipients;
	ApiKey          string                        `json:"-" xml:"-"` // api key used to identify client app;
	ServerUrl       string                        `json:"-" xml:"-"` // backend server url;
	ServerKeys      openpgp.EntityList            `json:"-" xml:"-"` // backend server keys;
	KeyserverKeys   openpgp.EntityList            `json:"-" xml:"-"` // the public keys of the server specified in the url;
	KeyserverUrl    string                        `json:"-" xml:"-"` // key servers.
}

// NewClient creates a new client struct starting minimum
// required data: keyserver url, 3nigm4 url and a keyring
// path address.
func NewClient(keyringPath string, keyserverUrl string, serverUrl string) *Client {
	// init client object
	c := Client{
		Sessions:        make([]messages.SessionKeys, 0),
		RecipientsCache: make(map[string]openpgp.EntityList),
		KeyserverUrl:    keyserverUrl,
		ServerUrl:       serverUrl,
	}

	return &c
}

// KeybaseStatusRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybaseStatusRes struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

// KeybaseBasicsRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybaseBasicsRes struct {
	Username      string `json:"username"`
	Ctime         uint   `json:"ctime"`
	Mtime         uint   `json:"mtime"`
	IdVersion     int    `json:"id_version"`
	TrackVersion  int    `json:"track_version"`
	LastIdChange  uint   `json:"last_id_change"`
	UsernameCased string `json:"username_cased"`
}

// KeybasePictureRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybasePictureRes struct {
	Url    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

// KeybasePicturesRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybasePicturesRes struct {
	Primary KeybasePictureRes `json:"primary"`
}

// KeybasePublicKeyRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybasePublicKeyRes struct {
	KeyFingerprint string `json:"key_fingerprint"`
	Kid            string `json:"kid"`
	SigningKid     string `json:"signing_kid"`
	KeyType        int    `json:"key_type"`
	Bundle         string `json:"bundle"`
	Mtime          uint   `json:"mtime"`
	Ctime          uint   `json:"ctime"`
	Ukbid          string `json:"ukbid"`
	KeyBits        uint   `json:"key_bits"`
	KeyAlgo        uint   `json:"key_algo"`
	KeyLevel       uint   `json:"key_level"`
	Status         int    `json:"status"`
	SelfSigned     bool   `json:"self_signed"`
	ExpirationTime uint   `json:"etime"`
}

// KeybaseProfileRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybaseProfileRes struct {
	Mtime    uint   `json:"mtime"`
	FullName string `json:"full_name"`
	Location string `json:"location"`
	Bio      string `json:"bio"`
}

// KeybasePublicKeysRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybasePublicKeysRes struct {
	Primary       KeybasePublicKeyRes `json:"primary"`
	PgpPublicKeys []string            `json:"pgp_public_keys"`
}

// KeybaseThemRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybaseThemRes struct {
	Id         string               `json:"id"`
	Basics     KeybaseBasicsRes     `json:"basics"`
	Pictures   KeybasePicturesRes   `json:"pictures"`
	PublicKeys KeybasePublicKeysRes `json:"public_keys"`
	CsrfToken  []byte               `json:"csrf_token"`
	Profile    KeybaseProfileRes    `json:"profile"`
}

// KeybaseUserLookupRes keybase response structures
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.
type KeybaseUserLookupRes struct {
	Status KeybaseStatusRes `json:"status"`
	Them   []KeybaseThemRes `json:"them"`
}

// Constant Keybase root paths.
const (
	RootLookupUrl = "_/api/1.0/user/lookup.json" // base URL path.
)

func (c *Client) composeKeyserverUrl() (*url.URL, error) {
	if len(c.KeyserverUrl) == 0 {
		return nil, fmt.Errorf("keyserver url must not be nil")
	}
	// compose url
	if c.KeyserverUrl[len(c.KeyserverUrl)-1] != '/' {
		c.KeyserverUrl = c.KeyserverUrl + "/"
	}
	u, err := url.Parse(c.KeyserverUrl + RootLookupUrl)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (c *Client) serverUrl(path string) (*url.URL, error) {
	if len(c.ServerUrl) == 0 {
		return nil, fmt.Errorf("server url must not be nil")
	}
	// compose url
	if c.ServerUrl[len(c.ServerUrl)-1] != '/' {
		c.ServerUrl = c.ServerUrl + "/"
	}
	u, err := url.Parse(c.ServerUrl + path)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetRecipientPublicKey uses Keybase APIs to retrieve recipient's
// PGP public key, if any.
func (c *Client) GetRecipientPublicKey(recipientIds []string) (*KeybaseUserLookupRes, error) {
	u, err := c.composeKeyserverUrl()
	if err != nil {
		return nil, err
	}
	// compose query
	q := u.Query()
	var recipients string
	for idx, recipient := range recipientIds {
		if idx != 0 &&
			idx != len(recipients)-1 {
			recipients = recipients + ","
		}
		recipients = recipients + recipient
	}
	q.Set("usernames", recipients)
	u.RawQuery = q.Encode()

	// request syncronously
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("unable to process the user lookup request, status: %d should be: %s (%s) cause: %s", resp.Status, strconv.Itoa(http.StatusOK), http.StatusText(http.StatusOK), string(body))
		return nil, err
	}
	// read body struct
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// unmarshal in structure
	var lr KeybaseUserLookupRes
	err = json.Unmarshal(body, &lr)
	if err != nil {
		return nil, err
	}

	return &lr, nil
}

// HandshakeMsg is the message used in
// handshake operations for the chat functionality
// this structure is used to exchange all session
// related keys.
type HandshakeMsg struct {
	SessionKeys []byte `json:"sessionk"`
	ServerKeys  []byte `json:"serverk"`
}

// PostNewSessionResp struct used to create
// a new session for chatting via REST APIs.
type PostNewSessionResp struct {
	SessionId []byte `json:"sessionid"`
}

// PostNewSession request the server to create a new session
// targeting some users.
func (c *Client) PostNewSession(session *messages.SessionKeys, recipients openpgp.EntityList, ttl uint64) (*PostNewSessionResp, error) {
	if session == nil ||
		len(recipients) == 0 {
		return nil, fmt.Errorf("invalid arguments, should not be nil")
	}
	// encrypt for recipients
	sessionenc, err := session.EncryptForRecipientsHandshake(recipients, c.PrivateKey)
	if err != nil {
		return nil, err
	}
	// encrypt for server
	serverenc, err := session.EncryptForServerHandshake(c.ServerKeys, c.PrivateKey, ttl)
	if err != nil {
		return nil, err
	}

	hm := HandshakeMsg{
		SessionKeys: sessionenc,
		ServerKeys:  serverenc,
	}

	jsonData, err := json.Marshal(hm)
	if err != nil {
		return nil, err
	}

	u, err := c.serverUrl("sessions")
	if err != nil {
		return nil, err
	}

	// enroll syncronously
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		bs, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("unable to process the enroll request, status:%d should be %d (%s) cause: %s", resp.Status, strconv.Itoa(http.StatusCreated), http.StatusText(http.StatusCreated), string(bs))
		return nil, err
	}
	// read response body
	body, _ := ioutil.ReadAll(resp.Body)
	var sessionResponse PostNewSessionResp
	err = json.Unmarshal(body, &sessionResponse)
	if err != nil {
		return nil, err
	}

	return &sessionResponse, nil
}

// MessageRequest structu represent a chat
// message.
type MessageRequest struct {
	SessionId   []byte `json:"sessionid"`
	MessageBody []byte `json:"messagebody"`
}

// MessageResponse is returned by the server
// when a message request is sent.
type MessageResponse struct {
	Counter uint64 `json:"counter"`
}

// PostMessage send a new chat message to the server and
// manage the produced response.
func (c *Client) PostMessage(session *messages.SessionKeys, message []byte) (uint64, error) {
	if len(message) == 0 ||
		session == nil {
		return 0, fmt.Errorf("unexpected arguments, should have not nil message and session")
	}

	// encrypt message
	encrypted, err := session.EncryptMessage(message, c.PrivateKey)
	if err != nil {
		return 0, fmt.Errorf("unable to encrypt the message, %s", err.Error())
	}

	// create message
	body := &MessageRequest{
		SessionId:   session.SessionId,
		MessageBody: encrypted,
	}
	// encode
	encoded, err := json.Marshal(body)

	u, err := c.serverUrl("messages")
	if err != nil {
		return 0, err
	}

	// enroll syncronously
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(encoded))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusCreated {
		bs, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("unable to process the message request, status:%d should be %d (%s) cause: %s", resp.Status, strconv.Itoa(http.StatusCreated), http.StatusText(http.StatusCreated), string(bs))
		return 0, err
	}
	// read response body
	response, _ := ioutil.ReadAll(resp.Body)
	var messageResponse MessageResponse
	err = json.Unmarshal(response, &messageResponse)
	if err != nil {
		return 0, err
	}

	return messageResponse.Counter, nil
}
