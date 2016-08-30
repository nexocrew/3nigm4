package resource

import (
	"net/http"
)

const kMuxChatName = "chat_name"

// Resource represents a generic REST resource.
type Resource interface{}

// GetSupported
type GetSupported interface {
	Get(*http.Request) (int, Resource)
}

// PostSupported
type PostSupported interface {
	Post(*http.Request) (int, Resource)
}

// PutSupported
type PutSupported interface {
	Put(*http.Request) (int, Resource)
}

// DeleteSupported
type DeleteSupported interface {
	Delete(*http.Request) (int, Resource)
}

// HeadSupported
type HeadSupported interface {
	Head(*http.Request) (int, Resource)
}

// PatchSupported
type PatchSupported interface {
	Patch(*http.Request) (int, Resource)
}

// ResourcePath defines a route pattern for a resource.
type ResourcePath struct {
	resource Resource
	pattern  string
}

func (r *ResourcePath) GetResource() Resource {
	return r.resource
}

func (r *ResourcePath) GetPattern() string {
	return r.pattern
}

// GetResources
// returns the route pattterns for all resources
func GetResources() []ResourcePath {
	return []ResourcePath{
		ResourcePath{
			new(Ping),
			"/ping",
		},
		ResourcePath{
			new(ChatCollection),
			"/chats",
		},
		ResourcePath{
			new(ChatResource),
			"/chat/{" + kMuxChatName + "}",
		},
		ResourcePath{
			new(MessagesCollection),
			"/chat/{" + kMuxChatName + "}/messages",
		},
		ResourcePath{
			new(MessagesCollection),
			"/chat/{" + kMuxChatName + "}/files",
		},
		ResourcePath{
			new(MessageResource),
			"/chat/{" + kMuxChatName + "}/message",
		},
		ResourcePath{
			new(MessageResource),
			"/chat/{" + kMuxChatName + "}/file",
		},
	}
}
