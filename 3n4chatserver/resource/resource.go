//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package resource

import (
	"github.com/nexocrew/3nigm4/lib/auth"
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

const muxChatName = "chat_name"
const usernameHttpHeader = "X-Username"
const passwordHttpHeader = "X-Hashed-Password"

// GetResources
// returns the route pattterns for all resources
func GetResources() []h.ResourcePath {
	return []h.ResourcePath{
		h.ResourcePath{
			new(Ping),
			"/ping",
		}, h.ResourcePath{
			new(Backdoor),
			"/backdoor",
		},
		h.ResourcePath{
			new(Auth),
			"/logout",
		},
		h.ResourcePath{
			new(ChatCollection),
			"/chats",
		},
		h.ResourcePath{
			new(ChatResource),
			"/chat/{" + muxChatName + "}",
		},
		h.ResourcePath{
			new(MessagesCollection),
			"/chat/{" + muxChatName + "}/messages",
		},
		h.ResourcePath{
			new(MessagesCollection),
			"/chat/{" + muxChatName + "}/files",
		},
		h.ResourcePath{
			new(MessageResource),
			"/chat/{" + muxChatName + "}/message",
		},
		h.ResourcePath{
			new(MessageResource),
			"/chat/{" + muxChatName + "}/file",
		},
	}
}

func GetLoginResource(authClient *auth.AuthRPC) []h.ResourcePath {
	return []h.ResourcePath{
		h.ResourcePath{
			&Auth{
				AuthClient: authClient,
			},
			"/login",
		},
	}
}
