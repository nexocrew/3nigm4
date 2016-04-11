//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package client

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

import (
	"github.com/nexocrew/3nigm4/lib/messages"
)

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

func NewClient(keyringPath string, keyserverUrl string) *Client {
	// init client object
	c := Client{
		Sessions:        make([]messages.SessionKeys, 0),
		RecipientsCache: make(map[string]openpgp.EntityList),
		KeyserverUrl:    keyserverUrl,
	}

	return &c
}

// Keybase response structures,
// see https://keybase.io/docs/api/1.0/call/user/lookup
// for details.

type KeybaseStatusRes struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

type KeybaseBasicsRes struct {
	Username      string `json:"username"`
	Ctime         uint   `json:"ctime"`
	Mtime         uint   `json:"mtime"`
	IdVersion     int    `json:"id_version"`
	TrackVersion  int    `json:"track_version"`
	LastIdChange  uint   `json:"last_id_change"`
	UsernameCased string `json:"username_cased"`
}

type KeybasePictureRes struct {
	Url    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

type KeybasePicturesRes struct {
	Primary KeybasePictureRes `json:"primary"`
}

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

type KeybaseProfileRes struct {
	Mtime    uint   `json:"mtime"`
	FullName string `json:"full_name"`
	Location string `json:"location"`
	Bio      string `json:"bio"`
}

type KeybasePublicKeysRes struct {
	Primary       KeybasePublicKeyRes `json:"primary"`
	PgpPublicKeys []string            `json:"pgp_public_keys"`
}

type KeybaseThemRes struct {
	Id         string               `json:"id"`
	Basics     KeybaseBasicsRes     `json:"basics"`
	Pictures   KeybasePicturesRes   `json:"pictures"`
	PublicKeys KeybasePublicKeysRes `json:"public_keys"`
	CsrfToken  []byte               `json:"csrf_token"`
	Profile    KeybaseProfileRes    `json:"profile"`
}

type KeybaseUserLookupRes struct {
	Status KeybaseStatusRes `json:"status"`
	Them   []KeybaseThemRes `json:"them"`
}

const (
	RootLookupUrl = "_/api/1.0/user/lookup.json"
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

// Handshake message
type HandshakeMsg struct {
	SessionKeys []byte `json:"sessionk"`
	ServerKeys  []byte `json:"serverk"`
}

// PostNewSession request the server to create a new session
// targeting some users.
func (c *Client) PostNewSession(session *messages.SessionKeys, recipients openpgp.EntityList, ttl uint64) error {
	if session == nil ||
		len(recipients) == 0 {
		return fmt.Errorf("invalid arguments, should not be nil")
	}
	// encrypt for recipients
	sessionenc, err := session.EncryptForRecipientsHandshake(recipients, c.PrivateKey)
	if err != nil {
		return err
	}
	// encrypt for server
	serverenc, err := session.EncryptForServerHandshake(c.ServerKeys, c.PrivateKey, ttl)
	if err != nil {
		return err
	}

	hm := HandshakeMsg{
		SessionKeys: sessionenc,
		ServerKeys:  serverenc,
	}

	jsonData, err := json.Marshal(hm)
	if err != nil {
		return err
	}

	u, err := c.serverUrl("sessions")
	if err != nil {
		return err
	}

	// enroll syncronously
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		bs, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("unable to process the enroll request, status:%d should be %d (%s) cause: %s", resp.Status, strconv.Itoa(http.StatusCreated), http.StatusText(http.StatusCreated), string(bs))
		return err
	}
	return nil
}
