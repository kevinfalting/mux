package mux

import (
	"fmt"
	"net/http"
	"strings"
)

type methodOption func(map[string]http.Handler)

// Methods will return a handler that will gate handlers by method for a path.
// If no OPTIONS handler was provided, one will be created.
func Methods(options ...methodOption) http.Handler {
	methodHandlers := map[string]http.Handler{}
	for _, opt := range options {
		opt(methodHandlers)
	}

	if _, ok := methodHandlers[http.MethodOptions]; !ok {
		var allowMethods []string
		for method := range methodHandlers {
			allowMethods = append(allowMethods, method)
		}

		allowValue := strings.Join(allowMethods, ", ")
		methodHandlers[http.MethodOptions] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Allow", allowValue)
			w.Header().Add("Access-Control-Allow-Methods", allowValue)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, ok := methodHandlers[r.Method]
		if !ok {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// WithMethod will register the handler against the http method
func WithMethod(method string, h http.Handler) methodOption {
	if len(method) == 0 {
		panic("method must not be empty")
	}

	if h == nil {
		panic("handler must not be nil")
	}

	return func(m map[string]http.Handler) {
		if _, ok := m[method]; ok {
			panic(fmt.Sprintf("method %q already registered", method))
		}
		m[method] = h
	}
}

// WithGET will register the handler against method GET
func WithGET(h http.Handler) methodOption {
	return WithMethod(http.MethodGet, h)
}

// WithPOST will register the handler against method POST
func WithPOST(h http.Handler) methodOption {
	return WithMethod(http.MethodPost, h)
}

// WithPUT will register the handler against method PUT
func WithPUT(h http.Handler) methodOption {
	return WithMethod(http.MethodPut, h)
}

// WithPATCH will register the handler against method PATCH
func WithPATCH(h http.Handler) methodOption {
	return WithMethod(http.MethodPatch, h)
}

// WithDELETE will register the handler against method DELETE
func WithDELETE(h http.Handler) methodOption {
	return WithMethod(http.MethodDelete, h)
}

// WithOPTIONS will register the handler against method OPTIONS. Provide if you
// need to use a custom OPTIONS handler for this path.
func WithOPTIONS(h http.Handler) methodOption {
	return WithMethod(http.MethodOptions, h)
}
