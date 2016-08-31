//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//
package main

// Go standard libraries
import (
	"net/http"
)

// 3n4 libraries
import (
	res "github.com/nexocrew/3nigm4/3n4chatserver/resource"
	h "github.com/nexocrew/3nigm4/lib/httphandler"
)

// Third party libraries
import (
	"github.com/gorilla/mux"
)

func router() *mux.Router {
	// create router
	r := mux.NewRouter()
	// define auth routes
	r.HandleFunc(basePath+"/authsession", login).Methods("POST")
	r.HandleFunc(basePath+"/authsession", logout).Methods("DELETE")

	// utility routes
	// r.HandleFunc(basePath+"/ping", ping).Methods("GET")
	// r.HandleFunc("/{"+kMuxVersion+"}/backdoor", backdoor).Methods("GET")

	// REST resources
	paths := res.GetResources()
	for _, path := range paths {
		r.HandleFunc(basePath+path.Pattern, h.Handler(path.Resource))
	}
	// root routes
	http.Handle("/", r)

	return r
}

func login(rw http.ResponseWriter, request *http.Request) {

}

func logout(rw http.ResponseWriter, request *http.Request) {

}
