//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package client

import (
	"golang.org/x/crypto/openpgp"
	"net/http"
	"net/url"
)

type Client struct {
	Sessions        []SessionKeys                 `json:"-" xml:"-"` // client opened sessions;
	PrivateKey      *openpgp.Entity               `json:"-" xml:"-"` // user's in memory private key;
	PublicKey       *openpgp.Entity               `json:"-" xml:"-"` // user's in memory public key;
	RecipientsCache map[string]openpgp.EntityList `json:"-" xml:"-"` // cached recipients;
	KeyserverUrl    string                        `json:"-" xml:"-"` // key servers.
}

func NewClient(keyringPath string, keyserverUrl string) *Client {
	// init client object
	c := Client{
		Sessions:       make([]SessionKeys, 0),
		RecipientCache: make(map[string]penpgp.EntityList),
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

func (c *Client) GetRecipientPublicKey(recipientIds []string) {
	// compose url
	if c.KeyserverUrl[len(c.KeyserverUrl)-1] != '/' {
		c.KeyserverUrl = c.KeyserverUrl + "/"
	}
	url := url.Parse(c.KeyserverUrl + RootLookupUrl)

	/*
		// encode JSON body
		bd, err := json.Marshal(enroll)
		if err != nil {
			return err
		}

		// POST https://<addr>:<port/api/v1/enroll
		url := c.Switchboard.Address + ":" + strconv.Itoa(c.Switchboard.Port) + "/api/v1/enroll"

		// enroll syncronously
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bd))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusCreated {
			bs, _ := ioutil.ReadAll(resp.Body)
			err = errors.New("unable to process the enroll request, status: '" + resp.Status + "' should be: '" + strconv.Itoa(http.StatusCreated) + " " + http.StatusText(http.StatusCreated) + "' cause:" + string(bs))
			return err
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var er enrollmentResonse
		err = json.Unmarshal(bs, &er)
		if err != nil {
			return err
		}
		c.Switchboard.Token = er.Token
		return nil
	*/
}
