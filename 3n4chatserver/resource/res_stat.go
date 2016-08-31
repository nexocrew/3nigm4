package resource

import (
	"net/http"
)

import (
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

type Ping struct{}
type Pong struct {
	Status string `json:"status"`
	Error  string `json:"err"`
}

func (p *Ping) Get(r *http.Request) (int, h.Resource) {
	return http.StatusOK, Pong{"ok", ""}
}
