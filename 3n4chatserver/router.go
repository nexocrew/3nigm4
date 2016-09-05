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
	paths := res.GetLoginResource(authClient)
	for _, path := range paths {
		r.HandleFunc(basePath+path.Pattern, h.Handler(path.Resource))
	}

	// REST resources
	paths = res.GetResources()
	for _, path := range paths {
		r.HandleFunc(basePath+path.Pattern, h.Handler(path.Resource))
	}
	// root routes
	http.Handle("/", r)

	return r
}
