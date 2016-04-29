//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package client

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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
	testKeys1And2PrivateHex = "9501d8044d3c5c10010400b1d13382944bd5aba23a4312968b5095d14f947f600eb478e14a6fcb16b0e0cac764884909c020bc495cfcc39a935387c661507bdb236a0612fb582cac3af9b29cc2c8c70090616c41b662f4da4c1201e195472eb7f4ae1ccbcbf9940fe21d985e379a5563dde5b9a23d35f1cfaa5790da3b79db26f23695107bfaca8e7b5bcd00110100010003ff4d91393b9a8e3430b14d6209df42f98dc927425b881f1209f319220841273a802a97c7bdb8b3a7740b3ab5866c4d1d308ad0d3a79bd1e883aacf1ac92dfe720285d10d08752a7efe3c609b1d00f17f2805b217be53999a7da7e493bfc3e9618fd17018991b8128aea70a05dbce30e4fbe626aa45775fa255dd9177aabf4df7cf0200c1ded12566e4bc2bb590455e5becfb2e2c9796482270a943343a7835de41080582c2be3caf5981aa838140e97afa40ad652a0b544f83eb1833b0957dce26e47b0200eacd6046741e9ce2ec5beb6fb5e6335457844fb09477f83b050a96be7da043e17f3a9523567ed40e7a521f818813a8b8a72209f1442844843ccc7eb9805442570200bdafe0438d97ac36e773c7162028d65844c4d463e2420aa2228c6e50dc2743c3d6c72d0d782a5173fe7be2169c8a9f4ef8a7cf3e37165e8c61b89c346cdc6c1799d2b41054657374204b6579203120285253412988b804130102002205024d3c5c10021b03060b090807030206150802090a0b0416020301021e01021780000a0910a34d7e18c20c31bbb5b304009cc45fe610b641a2c146331be94dade0a396e73ca725e1b25c21708d9cab46ecca5ccebc23055879df8f99eea39b377962a400f2ebdc36a7c99c333d74aeba346315137c3ff9d0a09b0273299090343048afb8107cf94cbd1400e3026f0ccac7ecebbc4d78588eb3e478fe2754d3ca664bcf3eac96ca4a6b0c8d7df5102f60f6b00200009d01d8044d3c5c10010400b201df61d67487301f11879d514f4248ade90c8f68c7af1284c161098de4c28c2850f1ec7b8e30f959793e571542ffc6532189409cb51c3d30dad78c4ad5165eda18b20d9826d8707d0f742e2ab492103a85bbd9ddf4f5720f6de7064feb0d39ee002219765bb07bcfb8b877f47abe270ddeda4f676108cecb6b9bb2ad484a4f00110100010003fd17a7490c22a79c59281fb7b20f5e6553ec0c1637ae382e8adaea295f50241037f8997cf42c1ce26417e015091451b15424b2c59eb8d4161b0975630408e394d3b00f88d4b4e18e2cc85e8251d4753a27c639c83f5ad4a571c4f19d7cd460b9b73c25ade730c99df09637bd173d8e3e981ac64432078263bb6dc30d3e974150dd0200d0ee05be3d4604d2146fb0457f31ba17c057560785aa804e8ca5530a7cd81d3440d0f4ba6851efcfd3954b7e68908fc0ba47f7ac37bf559c6c168b70d3a7c8cd0200da1c677c4bce06a068070f2b3733b0a714e88d62aa3f9a26c6f5216d48d5c2b5624144f3807c0df30be66b3268eeeca4df1fbded58faf49fc95dc3c35f134f8b01fd1396b6c0fc1b6c4f0eb8f5e44b8eace1e6073e20d0b8bc5385f86f1cf3f050f66af789f3ef1fc107b7f4421e19e0349c730c68f0a226981f4e889054fdb4dc149e8e889f04180102000905024d3c5c10021b0c000a0910a34d7e18c20c31bb1a03040085c8d62e16d05dc4e9dad64953c8a2eed8b6c12f92b1575eeaa6dcf7be9473dd5b24b37b6dffbb4e7c99ed1bd3cb11634be19b3e6e207bed7505c7ca111ccf47cb323bf1f8851eb6360e8034cbff8dd149993c959de89f8f77f38e7e98b8e3076323aa719328e2b408db5ec0d03936efd57422ba04f925cdc7b4c1af7590e40ab00200009501fe044d3c5c33010400b488c3e5f83f4d561f317817538d9d0397981e9aef1321ca68ebfae1cf8b7d388e19f4b5a24a82e2fbbf1c6c26557a6c5845307a03d815756f564ac7325b02bc83e87d5480a8fae848f07cb891f2d51ce7df83dcafdc12324517c86d472cc0ee10d47a68fd1d9ae49a6c19bbd36d82af597a0d88cc9c49de9df4e696fc1f0b5d0011010001fe030302e9030f3c783e14856063f16938530e148bc57a7aa3f3e4f90df9dceccdc779bc0835e1ad3d006e4a8d7b36d08b8e0de5a0d947254ecfbd22037e6572b426bcfdc517796b224b0036ff90bc574b5509bede85512f2eefb520fb4b02aa523ba739bff424a6fe81c5041f253f8d757e69a503d3563a104d0d49e9e890b9d0c26f96b55b743883b472caa7050c4acfd4a21f875bdf1258d88bd61224d303dc9df77f743137d51e6d5246b88c406780528fd9a3e15bab5452e5b93970d9dcc79f48b38651b9f15bfbcf6da452837e9cc70683d1bdca94507870f743e4ad902005812488dd342f836e72869afd00ce1850eea4cfa53ce10e3608e13d3c149394ee3cbd0e23d018fcbcb6e2ec5a1a22972d1d462ca05355d0d290dd2751e550d5efb38c6c89686344df64852bf4ff86638708f644e8ec6bd4af9b50d8541cb91891a431326ab2e332faa7ae86cfb6e0540aa63160c1e5cdd5a4add518b303fff0a20117c6bc77f7cfbaf36b04c865c6c2b42754657374204b6579203220285253412c20656e637279707465642070726976617465206b65792988b804130102002205024d3c5c33021b03060b090807030206150802090a0b0416020301021e01021780000a0910d4984f961e35246b98940400908a73b6a6169f700434f076c6c79015a49bee37130eaf23aaa3cfa9ce60bfe4acaa7bc95f1146ada5867e0079babb38804891f4f0b8ebca57a86b249dee786161a755b7a342e68ccf3f78ed6440a93a6626beb9a37aa66afcd4f888790cb4bb46d94a4ae3eb3d7d3e6b00f6bfec940303e89ec5b32a1eaaacce66497d539328b00200009d01fe044d3c5c33010400a4e913f9442abcc7f1804ccab27d2f787ffa592077ca935a8bb23165bd8d57576acac647cc596b2c3f814518cc8c82953c7a4478f32e0cf645630a5ba38d9618ef2bc3add69d459ae3dece5cab778938d988239f8c5ae437807075e06c828019959c644ff05ef6a5a1dab72227c98e3a040b0cf219026640698d7a13d8538a570011010001fe030302e9030f3c783e148560f936097339ae381d63116efcf802ff8b1c9360767db5219cc987375702a4123fd8657d3e22700f23f95020d1b261eda5257e9a72f9a918e8ef22dd5b3323ae03bbc1923dd224db988cadc16acc04b120a9f8b7e84da9716c53e0334d7b66586ddb9014df604b41be1e960dcfcbc96f4ed150a1a0dd070b9eb14276b9b6be413a769a75b519a53d3ecc0c220e85cd91ca354d57e7344517e64b43b6e29823cbd87eae26e2b2e78e6dedfbb76e3e9f77bcb844f9a8932eb3db2c3f9e44316e6f5d60e9e2a56e46b72abe6b06dc9a31cc63f10023d1f5e12d2a3ee93b675c96f504af0001220991c88db759e231b3320dcedf814dcf723fd9857e3d72d66a0f2af26950b915abdf56c1596f46a325bf17ad4810d3535fb02a259b247ac3dbd4cc3ecf9c51b6c07cebb009c1506fba0a89321ec8683e3fd009a6e551d50243e2d5092fefb3321083a4bad91320dc624bd6b5dddf93553e3d53924c05bfebec1fb4bd47e89a1a889f04180102000905024d3c5c33021b0c000a0910d4984f961e35246b26c703ff7ee29ef53bc1ae1ead533c408fa136db508434e233d6e62be621e031e5940bbd4c08142aed0f82217e7c3e1ec8de574bc06ccf3c36633be41ad78a9eacd209f861cae7b064100758545cc9dd83db71806dc1cfd5fb9ae5c7474bba0c19c44034ae61bae5eca379383339dece94ff56ff7aa44a582f3e5c38f45763af577c0934b0020000"
)

func readerFromHex(s string) ([]byte, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestLookupKeybaseUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, kKeybaseUserLookupResponse)
	}))
	defer ts.Close()

	c := NewClient("", ts.URL, "")
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

	c := NewClient("", ts.URL, "")
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

	// load tests public/private keys
	data, _ := readerFromHex(testKeys1And2PrivateHex)
	tmpkeyr, err := openpgp.ReadKeyRing(bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Unable to extract private key: %d.\n", err.Error())
	}
	if len(tmpkeyr) != 2 {
		t.Fatalf("Unexpected number of elements: having %d expecting 2.\n", len(tmpkeyr))
	}
	// add to keyring
	kr = append(kr, tmp...)
	kr = append(kr, tmpkeyr[0])
	if len(kr) != 3 {
		t.Fatalf("Unexpected number of keys in the keyring: having %d expecting 3.\n", len(kr))
	}

	// create a session
	sk, err := messages.NewSessionKeys(kCreatorId, []byte(kPreshared), []string{"illordlo", "dystonie"})
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}
	sk.UserId = "userA"
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
	enc, err := c.Sessions[0].EncryptForRecipientsHandshake(kr, signers[0])
	if err != nil {
		t.Fatalf("Unable to encrypt init message: %s.\n", err.Error())
	}

	// decrypt using pvkey
	initsk, err := messages.SessionFromEncryptedMsg(enc, tmpkeyr, []byte(kPreshared))
	if err != nil {
		t.Fatalf("Unable to access encrypted session: %s.\n", err.Error())
	}
	// check resulting session
	if bytes.Compare(c.Sessions[0].MainSymmetricKey, initsk.MainSymmetricKey) != 0 {
		t.Fatalf("Main keys are not equal: %s != %s.\n", base64.StdEncoding.EncodeToString(c.Sessions[0].MainSymmetricKey), base64.StdEncoding.EncodeToString(initsk.MainSymmetricKey))
	}
	if bytes.Compare(c.Sessions[0].ServerSymmetricKey, initsk.ServerSymmetricKey) != 0 {
		t.Fatalf("Server keys are not equal: %s != %s.\n", base64.StdEncoding.EncodeToString(c.Sessions[0].ServerSymmetricKey), base64.StdEncoding.EncodeToString(initsk.ServerSymmetricKey))
	}
	if bytes.Compare(c.Sessions[0].PreSharedKey, initsk.PreSharedKey) != 0 {
		t.Fatalf("Pre-shared keys are not equal: %s != %s.\n", base64.StdEncoding.EncodeToString(c.Sessions[0].PreSharedKey), base64.StdEncoding.EncodeToString(initsk.PreSharedKey))
	}
	if c.Sessions[0].CreatorId != initsk.CreatorId {
		t.Fatalf("CreatorId is not equal: %s != %s.\n", c.Sessions[0].CreatorId, initsk.CreatorId)
	}
	if c.Sessions[0].PreSharedFlag != initsk.PreSharedFlag {
		t.Fatalf("Preshared flag is not equal: %v != %v.\n", c.Sessions[0].PreSharedFlag, initsk.PreSharedFlag)
	}
	if c.Sessions[0].UserId == initsk.UserId {
		t.Fatalf("UserId must be manually setted should not be unmarshaled with the structure.\n")
	}
	initsk.UserId = "userB"

	// both users try to encrypt a message
	message := []byte(kPlainText)
	if len(message) == 0 {
		t.Fatalf("Message must not be nil.\n")
	}
	encrypted, err := c.Sessions[0].EncryptMessage(message, nil)
	if err != nil {
		t.Fatalf("Unable to encrypt message: %s.\n", err.Error())
	}
	decrypted, err := c.Sessions[0].DecryptMessage(encrypted, nil, false)
	if err != nil {
		t.Fatalf("Transmitted session keys are not able to decypt message: %s.\n", err.Error())
	}
	if bytes.Compare(message, decrypted.Body) != 0 {
		t.Fatalf("Different bodies.\n")
	}

}

func TestNewSession(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, kKeybaseUserLookupResponse)
	}))
	defer ts.Close()

	c := NewClient("", ts.URL, "")
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
	// add to keyring
	kr = append(kr, tmp...)

	tsp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) == 0 {
			t.Fatalf("Unexpected message body lenght, esxpecting not nil.\n")
		}
		t.Logf("Request body: %s.\n", string(body))
		fmt.Fprintf(w, "{\"status\":0, \"error\":\"\"}")
	}))
	defer ts.Close()

	cp := NewClient("", "", tsp.URL)
	if c == nil {
		t.Fatalf("Client must exist.\n")
	}

	// create a session
	sk, err := messages.NewSessionKeys(kCreatorId, []byte(kPreshared), []string{"illordlo", "dystonie"})
	if err != nil {
		t.Fatalf("Unable to create session keys: %s.\n", err.Error())
	}
	if sk == nil {
		t.Fatalf("Returned object must never be nil.\n")
	}

	err = cp.PostNewSession(sk, kr, 1000)
	if err != nil {
		t.Fatalf("Unable to post new session: %s.\n", err.Error())
	}
}
