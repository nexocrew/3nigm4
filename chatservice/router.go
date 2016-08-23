//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package main

// Golang std libs
import (
	"net/http"
)

// Third party dependencies
import (
	"github.com/gorilla/mux"
)

func router() mux.Router {
	// create router
	route := mux.NewRouter()
	// define auth routes
	route.HandleFunc("/{"+kMuxVersion+"}/authsession", login).Methods("POST")
	route.HandleFunc("/{"+kMuxVersion+"}/authsession", logout).Methods("DELETE")
	// get user informations
	route.HandleFunc("/{"+kMuxVersion+"}/me", getChat).Methods("GET")
	// get chat informations
	route.HandleFunc("/{"+kMuxVersion+"}/chats", getAllChats).Methods("GET")
	route.HandleFunc("/{"+kMuxVersion+"}/chat/{"+kMuxChatName+"}", getChatStats).Methods("GET")
	route.HandleFunc("/{"+kMuxVersion+"}/chat/{"+kMuxChatName+"}/messages", getChatMessages).Methods("GET")
	route.HandleFunc("/{"+kMuxVersion+"}/chat/{"+kMuxChatName+"}/files", getChatFiles).Methods("GET")
	// post messages to chat
	route.HandleFunc("/{"+kMuxVersion+"}/chat/{"+kMuxChatName+"}/message", postMessage).Methods("POST")
	route.HandleFunc("/{"+kMuxVersion+"}/chat/{"+kMuxChatName+"}/file", postFile).Methods("POST")
	// utility routes
	route.HandleFunc("/{"+kMuxVersion+"}/ping", ping).Methods("GET")
	route.HandleFunc("/{"+kMuxVerions+"}/backdoor", backdoor).Methods("GET")
	// root routes
	http.Handle("/", route)
}
