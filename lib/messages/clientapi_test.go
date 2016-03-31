//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/03/2016
//
package message

import (
	"net/http"
	"net/http/httptest"
)

func TestEnrollClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "{\"token\":\"%s\"}", kDefaultToken)
	}))
	defer ts.Close()
	/*
		cnf := defaultConfig("0.0.0.0", 7733, "", "127.0.0.1", 9775, "127.0.0.1", 9776)
		fmt.Printf("%s.\n", string(cnf))

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal("Unable to parse httptest server url.")
		}

		addr := strings.Split(u.Host, ":")
		a := u.Scheme + "://" + addr[0]
		p, _ := strconv.Atoi(addr[1])
		c := basicConfig()
		c.Switchboard.Address = a
		c.Switchboard.Port = p

		err = c.enrollClient()
		if err != nil {
			t.Fatalf("Unable connect correctly to server. Error:%s.\n", err.Error())
		}
		if c.Switchboard.Token != kDefaultToken {
			t.Fatalf("Ivalid token setted. Expecting:%s having:%s.\n", kDefaultToken, c.Switchboard.Token)
		}
	*/
}
