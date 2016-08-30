package datastructure

type Chat struct {
	ChatID         int
	Participants   []string
	NumberMessages int
}

type Message interface {
	GetID() int
	GetBody() []byte
}

type TextMessage struct {
	Id   int
	Body []byte
}
