//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package client

import (
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
	Username     string `json:"username"`
	Ctime        uint   `json:"ctime"`
	Mtime        uint   `json:"mtime"`
	IdVersion    int    `json:"id_version"`
	TrackVersion int    `json:"track_version"`
	LastIdChange uint   `json:"last_id_change"`
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
	KeyType        int    `json:"key_type"`
	Bundle         string `json:"bundle"`
	Mtime          uint   `json:"mtime"`
	Ctime          uint   `json:"ctime"`
	Ukbid          string `json:"ukbid"`
}

type KeybasePublicKeysRes struct {
	Primary KeybasePublicKeyRes `json:"primary"`
}

type KeybaseThemRes struct {
	Id         string               `json:"id"`
	Basics     KeybaseBasicsRes     `json:"basics"`
	Pictures   KeybasePicturesRes   `json:"pictures"`
	PublicKeys KeybasePublicKeysRes `json:"public_keys"`
	CsrfToken  []byte               `json:"csrf_token"`
}

type KeybaseUserLookupRes struct {
	Status KeybaseStatusRes `json:"status"`
	Them   []KeybaseThemRes `json:"them"`
}

const (
	RootLookupUrl = "_/api/1.0/user/lookup.json"
)

func (c *Client) GetRecipientPublicKey(recipientIds []string) (*KeybaseUserLookupRes, error) {
	if len(c.KeyserverUrl) == 0 {
		return nil, fmt.Errorf("keyserver url must not be nil")
	}
	// compose url
	if c.KeyserverUrl[len(c.KeyserverUrl)-1] != '/' {
		c.KeyserverUrl = c.KeyserverUrl + "/"
	}
	url, err := url.Parse(c.KeyserverUrl + RootLookupUrl)
	if err != nil {
		return nil, err
	}
	// compose query
	q := url.Query()
	var recipients string
	for idx, recipient := range recipientIds {
		if idx != 0 &&
			idx != len(recipients)-1 {
			recipients = recipients + ","
		}
		recipients = recipients + recipient
	}
	q.Set("usernames", recipients)
	url.RawQuery = q.Encode()

	fmt.Printf("Query: %s.\n", url.String())

	// request syncronously
	req, err := http.NewRequest("GET", url.String(), nil)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var lr KeybaseUserLookupRes
	err = json.Unmarshal(body, &lr)
	if err != nil {
		return nil, err
	}

	return &lr, nil
}
