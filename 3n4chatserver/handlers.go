//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package main

// Golang std libs
import (
	"encoding/json"
	"net/http"
)

// 3n4 libraries
import (
	res "github.com/nexocrew/3nigm4/3n4chatserver/resource"
)

// REST methods allowed.
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	HEAD   = "HEAD"
	PATCH  = "PATCH"
)

func login(w http.ResponseWriter, r *http.Request) {

}

func logout(w http.ResponseWriter, r *http.Request) {

}

func newHandler(resource res.Resource) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		var handler func(*http.Request) (int, res.Resource)

		// verify if method is supported
		switch request.Method {
		case GET:
			if resource, ok := resource.(res.GetSupported); ok {
				handler = resource.Get
			}
		case POST:
			if resource, ok := resource.(res.PostSupported); ok {
				handler = resource.Post
			}
		case PUT:
			if resource, ok := resource.(res.PutSupported); ok {
				handler = resource.Put
			}
		case DELETE:
			if resource, ok := resource.(res.DeleteSupported); ok {
				handler = resource.Delete
			}
		case HEAD:
			if resource, ok := resource.(res.HeadSupported); ok {
				handler = resource.Head
			}
		case PATCH:
			if resource, ok := resource.(res.PatchSupported); ok {
				handler = resource.Patch
			}
		}

		if handler == nil {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// execute handler function
		code, data := handler(request)

		content, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		// send successful response
		rw.WriteHeader(code)
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.Write(content)
	}
}
