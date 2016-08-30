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

// 3n4 libraries
import (
	res "github.com/nexocrew/3nigm4/3n4chatserver/resource"
)

// Third party dependencies
import (
	"github.com/gorilla/mux"
)

func router() *mux.Router {
	// create router
	r := mux.NewRouter()
	// define auth routes
	r.HandleFunc(kBasePath+"/authsession", login).Methods("POST")
	r.HandleFunc(kBasePath+"/authsession", logout).Methods("DELETE")

	// utility routes
	// r.HandleFunc(kBasePath+"/ping", ping).Methods("GET")
	// r.HandleFunc("/{"+kMuxVersion+"}/backdoor", backdoor).Methods("GET")

	// REST resources
	paths := res.GetResources()
	for _, path := range paths {
		r.HandleFunc(kBasePath+path.GetPattern(), newHandler(path.GetResource()))
	}
	// root routes
	http.Handle("/", r)

	return r
}
