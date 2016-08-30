package resource

import (
	"net/http"
)

type Ping struct{}
type Pong struct {
	Status string `json:"status"`
	Error  string `json:"err"`
}

func (p *Ping) Get(r *http.Request) (int, Resource) {
	return http.StatusOK, Pong{"ok", ""}
}
