//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package httphandler

// Go standard libraries
import (
	"encoding/json"
	"net/http"
)

// REST methods allowed.
const (
	mGET    = "GET"
	mPOST   = "POST"
	mPUT    = "PUT"
	mDELETE = "DELETE"
	mHEAD   = "HEAD"
	mPATCH  = "PATCH"
)

// Handler returns an http HandleFunc that verifies
// the called HTTP method and wether the Resource supports it.
// If the support is provided than the proper method is called,
// otherwise a proper error code is sent to the client
func Handler(resource Resource) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		if request.ParseForm() != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		var handler func(*http.Request) (int, Resource)

		// verify if method is supported
		switch request.Method {
		case mGET:
			if resource, ok := resource.(GetSupported); ok {
				handler = resource.Get
			}
		case mPOST:
			if resource, ok := resource.(PostSupported); ok {
				handler = resource.Post
			}
		case mPUT:
			if resource, ok := resource.(PutSupported); ok {
				handler = resource.Put
			}
		case mDELETE:
			if resource, ok := resource.(DeleteSupported); ok {
				handler = resource.Delete
			}
		case mHEAD:
			if resource, ok := resource.(HeadSupported); ok {
				handler = resource.Head
			}
		case mPATCH:
			if resource, ok := resource.(PatchSupported); ok {
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
