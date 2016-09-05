//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package resource

import (
	"net/http"
)

import (
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

type Backdoor struct{}
type Ping struct{}
type Pong struct {
	Status string `json:"status"`
	Error  string `json:"err"`
}

func (p *Ping) Get(r *http.Request) (int, h.Resource) {
	return http.StatusOK, Pong{"ok", ""}
}
