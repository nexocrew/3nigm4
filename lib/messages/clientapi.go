//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package message

import (
	"golang.org/x/crypto/openpgp"
	"net/http"
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

func (c *Client) GetRecipientPublicKey(recipientId string) {

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
