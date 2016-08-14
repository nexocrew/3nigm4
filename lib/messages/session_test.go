//
// 3nigm4 messages package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//

package messages

import (
	"bytes"
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
	kSenderId    = "golangtest@test.com"
	// pub   1024R/7F98BBCE 2014-01-04
	// uid   Golang Test (Private key password is 'golang') <golangtest@test.com>
	// sub   1024R/5F34A320 2014-01-04
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
	kPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

xsFNBFbVlNABEADGgfFMAkgrQeq97nuthMrbzKgQM38KHSEc/rfiDsrfJTaJZL7S
F8x43Ti01Uy51tLaiLubinXlRhdeChUdPmifb9LU0vRmHLZkPBo4ZsqnIDT8MeFX
IldoiGYaANXKkGoXvth6Wx25JKN2kY9xJ4xRPeM/c9H1os+joBaaSktzXGm+9VmD
1PlyOgp13dKuAtisXlZWHy2J+GT+M8Nu1swf2IX+3XQMsYalBopxjIdxMtNvrLOm
HsJTfWRq7RCYZaQ/4xXo4n8jpApc+U/LUxaXfsNAKqaQVEZKLfNWezkVXrWMXt9J
GowzeWpua79IabeiUm1Dciwk/MkjMhOrk/ageMmhsz2sM5Zp7bn7c/dqw+dU71ra
64HBh9QHZUTpLVdaSk/1joh54gtJlOvOTK/o6210raCDEM+59FwjVNMi2bdCqFXQ
mxL/G+LfF2mnnHPBeESNKh2f46IdbwAGwai5x00A6wZWf+RddV8yPzIge0Yn0RDz
hAu4FSoRJnAC1+uHIT5Uo6dWW8/5pmWKP68HpeFNmkX6Lx6C6muam9M5fbH9fx7L
cm+gXptEJPk8bNWkl5UBiP1sZEiiueT+TsjbIRwJAV3NFO3mAXqRz6KIfjNnP5yY
sEp+DOJq51xfTQR1V3l1sIQoWmHbziiyRlhcgwPBqIFn2LeshWHhGWOAzQARAQAB
zSRHdWlkbyBSb25jaGV0dGkgPGR5c3QwbmkzQGdtYWlsLmNvbT7CwXgEEwEIACwF
AlbVlNAJEN+70BU+rm7eAhsDBQkeEzgAAhkBBAsHCQMFFQgKAgMEFgABAgAA0yYQ
ALMnIX3klvxpUs3mAL9lH9aAR2eblPDbCXdM72WL7EHHGzfptkIjMHTDjj9gjn91
da0x/g1LRi1EjDsJjONC7cz0WpwkMPRW651h9mlqeBSg0RP75UvJ0j+jXrvbSc3I
SZOGWTQ+T15pJnvzRAa6d9xfV7Z9ka7AECYDghFh+0AAuVaBD5nivOeDTWf/RAzi
LcSql2KKSP63+kLGQ6nHOMiFzD05dcOcQQ/e44Fj/l4Qsw4ZZhnhLOpQZa6dVXUj
1KPJevuhtwiiqqHil+R15edn4b8GK/flHpapADB/uu+NCu3CyMctBp+YLmOZp8LD
Ipc/9SuHoVHmxuxL1ejcZZThTzZ0huFV+fDoXW3xMVny6jcvX6LYp2x7HtIK32AU
tywUF5FQgCYCWnL3gunMT5PLNjn6MM6S4wELBBRrPDfLbVpbYQbROamGfCiGk2G3
5AP8A4hS2SvA9AEk6ULpw+qOmerN/XmD+4VZidw8xDjZ28YI6ztjaURkJyoTEGoZ
vt+2PEH96tdvRDgeM6ZJqeW0mdjmD6FqiGxbmTA4U+dGNccMiLnGcYPdLDNzRixc
2xPRD/Thmxl4T2tTsPUhQ4f/Xh9ozIv8onJwU+6ZSn1CcFlWfwqXwtiwk2IIwzFz
S2D3Tlz8mKfTr0D9PDADh7S1KscEa0S9BAZBJcIHaWCOzRg8Z3VpZG8ucm9uY2hl
dHRpQGlrcy5pdD7CwXUEEwEIACkFAlbVlNAJEN+70BU+rm7eAhsDBQkeEzgABAsH
CQMFFQgKAgMEFgABAgAAc9QQAJFR0oGkRgiaFzwki8W3nxq3V7WRWv9qriOY9YKT
DGnNvVfGAUWx8KrwCp4LBXtLQcCFKwx2vGRNJsO+hyLlnxN+1/9P+ezY9wNhaBGK
cm14eLfcBf1ipt66P152gxA9yPOghUi1LY/V6zLuVviWt5uH9zEgWZNLZTwlNlNW
1TM4eQmx1UeLLMsZuH0KFWLrdDwQ3tGKDgu/FSaAgnNWW4wkhmmqxs/yYy+tR19P
nIZeCX2lNcVnEuDVJNQtGQ18+ABZm7BNkDBBDtML4+1SQkZNUmpmQoX/F2ZXFwDj
NEYCXtm480rEV5TKTx24nUeMJhbZSB34McE+Xv8IncD1ps1sxz6atwaAxaHXoL69
TIEZR7Yhdkslj5orNnXKClZtlWfCcXPrWIfqDm/tul6M3V2MKlL/0e+X/YH2TWDt
+lqwk0SXNj9GMgHOCBkT+bzyeSi3lBaSkCzurVYO8KW6Tpi/qp1ttuKrPau1WK04
B/8RzCXjsyCeMP14yk9Ba4gHsfcWfgDTYjVpcfMkqu67CqE3qN9oixIC7YSFHGyB
5jdZREgyoWIQS68rl3/jEp9jdHQuOgcMdp5rxdtc1FrCszc09RpCuWjv9+Q2yWOo
c6Fk5WuyQWmlliP2uEHqEvNkRbUHw/9yLzZvyv80XPCG86Os5GiZyAssJlPShfw5
hm4dzRs8Z3VpZG8ucm9uY2hldHRpQGR5c3RvLm9yZz7CwXUEEwEIACkFAlbVlNAJ
EN+70BU+rm7eAhsDBQkeEzgABAsHCQMFFQgKAgMEFgABAgAAqV8QAEsWJ7CNeb+D
80TzzXKy0WpdvV7xo0FEzeuJoBuah1vGuipVQ7x4Y+0ByUr4+jCv0V+OR9HnWbxt
bujtE5s258U1TrMmTFM8KNCDThPx7IxBPiO+Vjf+i1wgCp/Hw7olcNuprlG/ti9s
6qXOrljgCE6yYT3rx/5cCvEDEdXaoc2hsx6sGuA+mkP1CQMUQxpd8zJqTM8rldxc
DXdDoVcrUj5RELsJ+MFVUC+pLq2JKiO+w3R/yyN5A3IvfdkJOV75yKOoYWN0771p
6tiDdaDoCPvAxUw5j/TrBXhLXMtvUHsmoaHWOvIHg+xhWzdUlHjrIRWhjpwjwwuW
Jv0VWfoAN1UFeOR6Hl9S85NMBpuM9InQF83nDZXB7QuLcZatUIPrZ2qeHQLwwfby
XQKuUF1/t0b13WhF/6he7zxNke+Ms7UTuKntWMb1OFKIt+R3dmT8MRdG1ePUyNhQ
TDs+SurmnSIFdPR/3vjcoshaR4BYlFR6t+QptvNvtmmpJijj+XD1jNW0f4b09cSn
c6i+AH/dMYOzIK2Aamji2UsKANOjR5Yof6fzqvXFILhYzivo4umDQeRN52BKgzNg
/LY0NCiZFQDuPku6pn2u3IRnBB7cF9vP6+osE3Nwgn1XQIeFGVFXowT9DJqUF8T2
SGtZ5SJiFt3bmCZ1dBnTGhaNA+G0EVUFzsFNBFbVlNABEACgJO6QxtqUalhf4Dde
cSEZbhYcMRJJVhAfJKSROt3fPOPwJWh1ZuoMDAa1dC3Aa/h1qASkpTtpbILm2lDq
fmgkGJXrKDn8Yw6kFSBxXRzmANOpOJ+ZjSMm1b/RITXBt9JXcNOGz4vCsMxHGmTP
gpnGa5xbrM9WIQAwoT6+FmcCrh2D9GJzpJSDqqTFGS/J4D/AWEEe1itBqk09JBZM
XX1ROKIoap8NJ25/CNpXhoKqML7Jj8Uab9sBjKTN/Yb0lC7371V1/MvcFKGj+9hn
KhFgcFvnAxtnAj8GhtsCPZMgvP3FUOYdakbi+gVNJBZrAatA5qCLFDoxs3tgaASE
17obHaR0rFMAJ0PqZRKKFhCJnEWzG59tWXV+fCf8/bsXTt1cogNrTyUmf1Yq9fNt
Ee31ts/UrfY/CAORytw0fUdf22RX2hnuuRtRcu/jkrOtkUlR641r6Asc+bHpzd3I
RlULPWRyibMJy1H2SwlJyFrtlF0cmaOSGhy4mqc8MeW5LGT25x6NxrH7T5eQmAsa
WcsfFhOOctZcqmRGC5cwFSsPymNVF1XrbKHnEp1XUCWxmWE/Ty4eOZmO18ipGVF/
XNinTK6cZdXevW1zXl4xS8dkYHmf5P1f5qTAJ8jZRMvu7aPwDUF4nRXY/BKaoFQU
Cr5EFUOqLuSToyfIGMxgeFUFOQARAQABwsF1BBgBCAApBQJW1ZTQCRDfu9AVPq5u
3gIbDAUJHhM4AAQLBwkDBRUICgIDBBYAAQIAAA9+EACBLh2hJSRK+PVqHSWP+W2S
KOeamWiVCaFXGUFtJk2tJNDT3ClK60ENO9woNbf+KE4V1uGQd5opg5JlTsOvXbcW
AQDX5J7Vwx/+3o/6fwWXpM+Kze8L9fAap3ntZ/yFnye5dgxEqbrVikhBLKZZoSvV
qRLDTUBjMwlq+kPa2LTDXW3Fc+dyakDVNn08flVoqgVHVgT6YQhw97JQi/UPZN0I
CEL4xRHdIhzv4q+awjiT/TQJkure+zVuYm+EIAp1O9NoxUw9I1R3JL74M1mbRU83
iAzUKOEISpcZRV3i693va2tWcZaTpejh18/xMWeEtQS1KcaztN6V+ddstNhongoc
OrccvNCbwIsxK1h2tlCr05dIi3EQMlZLwamYf+OZXENI+6u/I47bnJSjJl9Gwsat
elyuXmuZo/1QaWdbaxQyyEdOdk7+hHzfXE2sIAdg4x3baxfT7qfXI6zqpLWo6vIV
y7rRYOGsoVkP648H5J5TIikfAd368/xFrDPHXYr7bA8KR3WfD2SI0YPuhVqqD+jb
0ZenJ7x+CpT2H9AS7FvspTPIwFyAj+EuCj84Sy6Nu7vUbu4EjHiF1w/eSvuAH+oA
cLbzULoSZctjW9I93SolwBTVxwWJKgvMzXe6eDZ6rjUjSukowxprX7nsk+2WDrQq
lw1nIDx1uwkeAfXXcViFBg==
=Ifs7
-----END PGP PUBLIC KEY BLOCK-----`
)

func TestEncryptForRecipients(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared), []string{})
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}
	// get entity
	signers, err := crypto3n.ReadArmoredKeyRing([]byte(kPrivateKey), []byte("golang"))
	if err != nil {
		t.Fatalf("Unable to read private key: %s.\n", err.Error())
	}
	if len(signers) == 0 {
		t.Fatalf("Expecting at least 1 entity having %d.\n", len(signers))
	}
	entityList, err := crypto3n.ReadArmoredKeyRing([]byte(kPublicKey), nil)
	if err != nil {
		t.Fatalf("Unable to access public armored key: %s.\n", err.Error())
	}
	handshake, err := sk.EncryptForRecipientsHandshake(entityList, signers[0])
	if err != nil {
		t.Fatalf("Unable to produce a valid handshake message: %s.\n", err.Error())
	}
	if len(handshake) == 0 {
		t.Fatalf("Unexpected handshake length should be not nil.\n")
	}
}

func TestEncryptForServer(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared), []string{"illordlo", "dystonie"})
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}
	// get entity
	signers, err := crypto3n.ReadArmoredKeyRing([]byte(kPrivateKey), []byte("golang"))
	if err != nil {
		t.Fatalf("Unable to read private key: %s.\n", err.Error())
	}
	if len(signers) == 0 {
		t.Fatalf("Expecting at least 1 entity having %d.\n", len(signers))
	}
	entityList, err := crypto3n.ReadArmoredKeyRing([]byte(kPublicKey), nil)
	if err != nil {
		t.Fatalf("Unable to access public armored key: %s.\n", err.Error())
	}
	handshake, err := sk.EncryptForServerHandshake(entityList, signers[0], 2000)
	if err != nil {
		t.Fatalf("Unable to produce a valid server handshake message: %s.\n", err.Error())
	}
	if len(handshake) == 0 {
		t.Fatalf("Unexpected handshake length should be not nil.\n")
	}
}

func TestEncryptMessageNotSigned(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared), []string{})
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
	sk.ServerTmpKey = []byte(kServerKey)
	sk.UserId = kSenderId
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

	// check message struct
	if len(message.Signature) == 0 {
		t.Fatalf("Signature must be not empty.\n")
	}
	if len(message.Message.EncryptedBody) == 0 {
		t.Fatalf("Encrypted body should be not empty.\n")
	}
	if len(message.Message.EncryptedSenderId) == 0 {
		t.Fatalf("Encrypted sender id should not be nil.\n")
	}
	if len(message.Message.SenderId) != 0 {
		t.Fatalf("Plain text sender id should not be setted in encrypted structure.\n")
	}
	if len(message.Message.Body) != 0 {
		t.Fatalf("Plain text body should not be setted in encrypted structure.\n")
	}
	if bytes.Compare(message.Message.SessionId, []byte(kSessionName)) != 0 {
		t.Fatalf("Session id is different from expected %s, having %s.\n", kSessionName, string(message.Message.SenderId))
	}

	// decrypt message
	decrypted, err := sk.DecryptMessage(encrypted, nil, false)
	if err != nil {
		t.Fatalf("Unable to decrypt message: %s.\n", err.Error())
	}
	t.Logf("Decrypted: %s.\n", string(decrypted.Body))

	if bytes.Compare(decrypted.Body, []byte(kTestMessage)) != 0 {
		t.Fatalf("Decrypted data is different from the original message.\n")
	}
	if decrypted.SenderId != kSenderId {
		t.Fatalf("Unexpected sender, having %s expecting %s.\n", decrypted.SenderId, kSenderId)
	}
}

func TestEncryptMessageSigned(t *testing.T) {
	sk, err := NewSessionKeys(kCreatorId, []byte(kPreshared), []string{})
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
	sk.UserId = kSenderId
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

	// check message struct
	if len(message.Signature) == 0 {
		t.Fatalf("Signature must be not empty.\n")
	}
	if len(message.Message.EncryptedBody) == 0 {
		t.Fatalf("Encrypted body should be not empty.\n")
	}
	if len(message.Message.EncryptedSenderId) == 0 {
		t.Fatalf("Encrypted sender id should not be nil.\n")
	}
	if len(message.Message.SenderId) != 0 {
		t.Fatalf("Plain text sender id should not be setted in encrypted structure.\n")
	}
	if len(message.Message.Body) != 0 {
		t.Fatalf("Plain text body should not be setted in encrypted structure.\n")
	}
	if bytes.Compare(message.Message.SessionId, []byte(kSessionName)) != 0 {
		t.Fatalf("Session id is different from expected %s, having %s.\n", kSessionName, string(message.Message.SenderId))
	}

	// decrypt message
	decrypted, err := sk.DecryptMessage(encrypted, entities, true)
	if err != nil {
		t.Fatalf("Unable to decrypt message: %s.\n", err.Error())
	}
	t.Logf("Decrypted: %s.\n", string(decrypted.Body))

	if bytes.Compare(decrypted.Body, []byte(kTestMessage)) != 0 {
		t.Fatalf("Decrypted data is different from the original message.\n")
	}
	if decrypted.SenderId != kSenderId {
		t.Fatalf("Unexpected sender, having %s expecting %s.\n", decrypted.SenderId, kSenderId)
	}
}
