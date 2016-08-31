package resource

import (
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

const kMuxChatName = "chat_name"

// GetResources
// returns the route pattterns for all resources
func GetResources() []h.ResourcePath {
	return []h.ResourcePath{
		h.ResourcePath{
			new(Ping),
			"/ping",
		},
		h.ResourcePath{
			new(ChatCollection),
			"/chats",
		},
		h.ResourcePath{
			new(ChatResource),
			"/chat/{" + kMuxChatName + "}",
		},
		h.ResourcePath{
			new(MessagesCollection),
			"/chat/{" + kMuxChatName + "}/messages",
		},
		h.ResourcePath{
			new(MessagesCollection),
			"/chat/{" + kMuxChatName + "}/files",
		},
		h.ResourcePath{
			new(MessageResource),
			"/chat/{" + kMuxChatName + "}/message",
		},
		h.ResourcePath{
			new(MessageResource),
			"/chat/{" + kMuxChatName + "}/file",
		},
	}
}
