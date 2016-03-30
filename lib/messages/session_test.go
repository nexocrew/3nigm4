//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//
package message

import (
	"encoding/json"
	"testing"
)

import (
	crypto3n "github.com/nexocrew/3nigm4/lib/crypto"
)

const (
	kCreatorId   = "user.test@mail.com"
	kPreshared   = "presharedsec"
	kTestMessage = "This is my super-secret message that should never be intercepted and readed by anyone."
	kSessionName = "000000000001"
	kServerKey   = "serverkey000001"
	kPrivateKey  = `-----BEGIN PGP PRIVATE KEY BLOCK-----
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

func TestEncryptMessage(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared))
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}
	// get entity
	entities, err := crypto3n.ReadArmoredKeyRing([]byte(kPrivateKey), []byte("golang"))
	if err != nil {
		t.Fatalf("Unable to read private key: %s.\n", err.Error())
	}
	if len(entities) == 0 {
		t.Fatalf("Expecting at least 1 entity having %d.\n", len(entities))
	}
	// set manually elements
	sk.SessionId = []byte(kSessionName)
	sk.ServerSymmetricKey = []byte(kServerKey)
	encrypted, err := sk.EncryptMessage([]byte(kTestMessage), entities[0])
	if err != nil {
		t.Fatalf("Unable to encrypt the message: %s.\n", err.Error())
	}
	t.Logf("Message: %s.\n", string(encrypted))

	// decode message
	var message SignedMessage
	err = json.Unmarshal(encrypted, &message)
	if err != nil {
		t.Fatalf("Unexpected object type unable to unmarshal: %s.\n", err.Error())
	}

	// decrypt message
	decrypted, _, _, err := sk.DecryptMessage(encrypted, nil, false)
	if err != nil {
		t.Fatalf("Unable to decrypt message: %s.\n", err.Error())
	}
	t.Logf("Decrypted: %s.\n", string(decrypted))

}
