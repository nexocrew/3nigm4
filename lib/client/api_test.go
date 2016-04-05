//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/base64"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
	"github.com/nexocrew/3nigm4/lib/messages"
)

const (
	kCreatorId = "user.test@mail.com"
	kPreshared = "presharedsec"
	kPlainText = `This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted. 
	This is a test message, it will be encrypted.`
	kPrivateKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1
 
lQH+BFLHbYYBBADCjgKHmPmwBxI3c3DPVoSdu0+EJl/EsS2HEaN63dnLkGsMAs+4
32wsywmMrzKqCL40sbhJVYBcfe0chL+cry4O54DX7+gA0ZSVzFUN2EGocnkaHzyS
fuUtBdCTmoWZZAGFiBwlIS7aE/86SOyHksFo8LRC9W/GIWQS2PbcadvUywARAQAB
/gMDApJxOwcsfChBYCCmhOAvotKdYcy7nuG7dyGDBlpclLJtH/PaakKSE33NtEj4
1fyixQOdwApxvuQ2P0VX3pie/De1KpbeqXfnPLsmsXQwrRPOo38T5zeJ5ToWUGDC
Oia69ep3kmHbAW41EBH/uk/nMM91QUdl4mkYsc3dhVOXbmf0xyRoP/Afqha4UhdZ
0XKlIZP1a5+3NF/Q6dAVG0+FlO5Hcai8n98jW0id8Yf6zI+1gFGvYYKhlifkdJeK
Nf4YEvOXALEvaQqkcJOxEca+BmqsgCIFctJe9Bahx97Ep5hP7AH0aBmtZfmGmZwB
GYoevUtKa4ASVmK8RaddBvIjcrWsoAsYMpDGYaE0fcdtxsBf3uT1Q8IMsT+ZRjjV
TfvJ8aW14ZrLI98KdtXaOPZs91mML+3iw1c/1O/IEJfwxrUni2p/fDmCYU9eHR3u
Q0PwVR0MCUHI1fGuUoetW2gYIxfklvBtEFWW1BD6fCpCtERHb2xhbmcgVGVzdCAo
UHJpdmF0ZSBrZXkgcGFzc3dvcmQgaXMgJ2dvbGFuZycpIDxnb2xhbmd0ZXN0QHRl
c3QuY29tPoi4BBMBAgAiBQJSx22GAhsDBgsJCAcDAgYVCAIJCgsEFgIDAQIeAQIX
gAAKCRBVSiCHf5i7zqKJA/sFUM2TfL2VZKWC7E1N1wwZctB9Bf77SeAPSVpGCZ0c
iUYIFdwwGowKtjoDrsbYgPp+UGOyYMD6tGzWKaJrQQoDyaQqVVRhbNXB7Jz7JT2a
qKHD1t7cx5FfUzDMBNou3TOWHomDXyQGDAULAZnjaOj8/pDe6poxyBluSjMJUzfD
pp0B/gRSx22GAQQArUMDqkGng9Cppk73UBWBd7jhhbtk0eaRQh/goUHhKJerZ4LM
Q21IKyIX+GQbscDpccpXMI6eThXxrL+D8G4cNb4ewvT0zc20+T91ztgT9A/4Vifc
EPQCErTqY/oZphAzZM1p6sRenc22e42iT0Iibd5gCs2wnSNeUzybDcuQi2EAEQEA
Af4DAwKScTsHLHwoQWCYayWqio8purPTonYogZSN3QwaheS2Y0NE7skdLOvP97vi
Rh7BktS6Dkgu0T3D39+q0O6ZO7XErvTVoas1F0HXzId4tiIicmx4tYNyWI4NrSO7
6TQPz/bQe8ZN+plG5cgZowts6g6RSfQxoW21LrP8Lh+OEdcYwWf7BTukAYmD3oq9
RxdfYI7hnbVGFdOqQUQNcxZkbdrsF9ITjQb/KRln5/99E1Kp1D45VpPOs7NT3orA
mnfSslJXVNm1uK6FDBX2iUe3JaAmgh+RLGXQXRZKJW4DGDTyYdwR4hO8cYix2+8z
+XuwdVDPKBnzKn190m6xpdLyvKfj1BQhX14NShPQZ3QJiMU0k4Js23XSsWs9NSxI
FjjE9/mOFVUH25KN+X7rzBPo2S0pMQLqyQxSLIdI2LPDxzlknctT6OoBPKPJjb7S
Lt5GhIA5Cz+cohfX6LePG4FkvwU32tTRBz5YNhFBizmS+YifBBgBAgAJBQJSx22G
AhsMAAoJEFVKIId/mLvOulED/2uUh/qjOT468XoK6Xt837w45JQPpLqiGH9KJgqF
rUxJMw1bIE2G606OY6hCgeE+YC8qny29hQtXhKIquUI/0A1qK3aCZhwqyqT+QjvF
6Xi0i/HrgQwCyBopY3uGndMbvthxU0KO0d6seMZltHDr8YaU1JvDwNFDQVuw+Rqy
57ET
=nvLl
-----END PGP PRIVATE KEY BLOCK-----`
)

func TestLookupKeybaseUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, kKeybaseUserLookupResponse)
	}))
	defer ts.Close()

	c := NewClient("", ts.URL)
	if c == nil {
		t.Fatalf("Client must exist.\n")
	}

	res, err := c.GetRecipientPublicKey([]string{"dystonie", "illordlo"})
	if err != nil {
		t.Fatalf("Unable to process GET request: %s.\n", err.Error())
	}
	if res.Status.Code != 0 ||
		res.Status.Name != "OK" {
		t.Fatalf("Unexpected unmarshaling expecting \"OK\" having %s.\n", res.Status.Name)
	}
	if len(res.Them) != 2 {
		t.Fatalf("Expected 2 results, having %d.\n", len(res.Them))
	}

	dystonie := res.Them[0]
	illordlo := res.Them[1]
	// check basics
	if dystonie.Basics.Username != "dystonie" {
		t.Fatalf("Unexpected username having %s expecting \"dystonie\".\n", dystonie.Basics.Username)
	}
	if illordlo.Basics.Username != "illordlo" {
		t.Fatalf("Unexpected username having %s expecting \"illordlo\".\n", illordlo.Basics.Username)
	}
	if dystonie.Basics.Ctime == 0 {
		t.Fatalf("Unexpected creation time should be > 0 having %d.\n", dystonie.Basics.Ctime)
	}
	if illordlo.Basics.Ctime == 0 {
		t.Fatalf("Unexpected creation time should be > 0 having %d.\n", illordlo.Basics.Ctime)
	}
	if dystonie.Basics.Mtime == 0 {
		t.Fatalf("Unexpected modification time should be > 0 having %d.\n", dystonie.Basics.Mtime)
	}
	if illordlo.Basics.Mtime == 0 {
		t.Fatalf("Unexpected modification time should be > 0 having %d.\n", illordlo.Basics.Mtime)
	}
	if dystonie.Basics.IdVersion == 0 {
		t.Fatalf("Unexpected id version should be > 0 having %d.\n", dystonie.Basics.IdVersion)
	}
	if illordlo.Basics.IdVersion == 0 {
		t.Fatalf("Unexpected id version should be > 0 having %d.\n", illordlo.Basics.IdVersion)
	}
	if dystonie.Basics.LastIdChange == 0 {
		t.Fatalf("Unexpected last id changed should be > 0 having %d.\n", dystonie.Basics.LastIdChange)
	}
	if illordlo.Basics.LastIdChange == 0 {
		t.Fatalf("Unexpected last id changed should be > 0 having %d.\n", illordlo.Basics.LastIdChange)
	}
	if dystonie.Basics.UsernameCased != "dystonie" {
		t.Fatalf("Unexpected username cased having %s expecting \"dystonie\".\n", dystonie.Basics.UsernameCased)
	}
	if illordlo.Basics.UsernameCased != "illordlo" {
		t.Fatalf("Unexpected username cased having %s expecting \"illordlo\".\n", illordlo.Basics.UsernameCased)
	}

	// check profile
	if dystonie.Profile.FullName != "Guido Ronchetti" {
		t.Fatalf("Unexpected full name having %s expecting \"Guido Ronchetti\".\n", dystonie.Profile.FullName)
	}
	if dystonie.Profile.Location != "Bombay City" {
		t.Fatalf("Unexpected location having %s expecting \"Bombay City\".\n", dystonie.Profile.Location)
	}
	if dystonie.Profile.Bio == "" {
		t.Fatalf("Unexpected bio having %s expecting not empty.\n", dystonie.Profile.Bio)
	}
	if dystonie.Profile.Mtime == 0 {
		t.Fatalf("Unexpected modification time should be > 0 having %d.\n", dystonie.Profile.Mtime)
	}

	// check public key
	if dystonie.PublicKeys.Primary.Bundle == "" {
		t.Fatalf("Nil public key should be valid.\n")
	}
	if dystonie.PublicKeys.Primary.KeyType != 1 {
		t.Fatalf("Unexpected key type: having %d expecting 1.\n", dystonie.PublicKeys.Primary.KeyType)
	}
	if dystonie.PublicKeys.Primary.SelfSigned != true {
		t.Fatalf("Unexpected key having signed and should be self signed.\n")
	}
	if dystonie.PublicKeys.Primary.ExpirationTime == 0 {
		t.Fatalf("Unexpected expiration time should not be 0.\n")
	}
}

func TestClientFlow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, kKeybaseUserLookupResponse)
	}))
	defer ts.Close()

	c := NewClient("", ts.URL)
	if c == nil {
		t.Fatalf("Client must exist.\n")
	}

	res, err := c.GetRecipientPublicKey([]string{"dystonie", "illordlo"})
	if err != nil {
		t.Fatalf("Unable to process GET request: %s.\n", err.Error())
	}
	if res.Status.Code != 0 ||
		res.Status.Name != "OK" {
		t.Fatalf("Unexpected unmarshaling expecting \"OK\" having %s.\n", res.Status.Name)
	}
	if len(res.Them) != 2 {
		t.Fatalf("Expected 2 results, having %d.\n", len(res.Them))
	}

	dystonie := res.Them[0]
	illordlo := res.Them[1]

	// load key in client cache
	kr, err := crypto3n.ReadArmoredKeyRing([]byte(dystonie.PublicKeys.Primary.Bundle), nil)
	if err != nil {
		t.Fatalf("Unable to read dystonie armored keyring: %s.\n", err.Error())
	}
	tmp, err := crypto3n.ReadArmoredKeyRing([]byte(illordlo.PublicKeys.Primary.Bundle), nil)
	if err != nil {
		t.Fatalf("Unable to read illordlo armored keyring: %s.\n", err.Error())
	}
	kr = append(kr, tmp...)

	// create a session
	sk, err := messages.NewSessionKeys(kCreatorId, []byte(kPreshared))
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}
	// add to client
	c.Sessions = append(c.Sessions, *sk)

	if len(c.Sessions) != 1 {
		t.Fatalf("Unexpected number of session per client: having %d expecting 1.\n", len(c.Sessions))
	}

	signers, err := crypto3n.ReadArmoredKeyRing([]byte(kPrivateKey), []byte("golang"))
	if err != nil {
		t.Fatalf("Unable to read private key: %s.\n", err.Error())
	}
	if len(signers) == 0 {
		t.Fatalf("Expecting at least 1 entity having %d.\n", len(signers))
	}
	// encrypted session initialisation
	enc, err := c.Sessions[0].EncryptForRecipients(kr, signers[0])
	if err != nil {
		t.Fatalf("Unable to encrypt init message: %s.\n", err.Error())
	}
	t.Logf("Init message: %s.\n", base64.StdEncoding.EncodeToString(enc))

}
