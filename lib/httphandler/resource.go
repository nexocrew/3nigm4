package httphandler

// Go standard libraries
import (
	"net/http"
)

// Resource represents a generic REST resource.
type Resource interface{}

// GetSupported inteface used to verify
// wether GET method is supported
type GetSupported interface {
	Get(*http.Request) (int, Resource)
}

// PostSupported inteface used to verify
// wether POST method is supported
type PostSupported interface {
	Post(*http.Request) (int, Resource)
}

// PutSupported inteface used to verify
// wether PUT method is supported
type PutSupported interface {
	Put(*http.Request) (int, Resource)
}

// DeleteSupported inteface used to verify
// wether DELETE method is supported
type DeleteSupported interface {
	Delete(*http.Request) (int, Resource)
}

// HeadSupported inteface used to verify
// wether HEAD method is supported
type HeadSupported interface {
	Head(*http.Request) (int, Resource)
}

// PatchSupported inteface used to verify
// wether PATCH method is supported
type PatchSupported interface {
	Patch(*http.Request) (int, Resource)
}

// ResourcePath defines a route pattern for a resource.
type ResourcePath struct {
	Resource Resource
	Pattern  string
}
