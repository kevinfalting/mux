package mux

import (
	"net/http"
	"strings"
)

// Mux wraps the http.ServeMux and provides a mechanism for registering
// middleware
type Mux struct {
	mux *http.ServeMux
	mw  []Middleware
}

// New will return an instance of a new Mux. The provided middleware will wrap
// every handler registered to the Mux.
func New(mw ...Middleware) *Mux {
	return &Mux{
		mux: http.NewServeMux(),
		mw:  mw,
	}
}

// ServeHTTP satisfies the handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Handle will register the provided handler on the mux, wrapped in the provided
// middleware(s). Middleware is envoked from left to right per request, after
// any mux level middleware.
func (m *Mux) Handle(pattern string, handler http.Handler, mw ...Middleware) {
	// handler specific middleware
	handler = WrapMiddleware(mw, handler)

	// mux middleware
	handler = WrapMiddleware(m.mw, handler)

	m.mux.Handle(pattern, handler)
}

// HandleFunc will register the provided handler function on the mux, wrapped in
// the provided middleware(s). Middleware is envoked from left to right per
// request, after any mux level middleware.
func (m *Mux) HandleFunc(pattern string, handler http.HandlerFunc, mw ...Middleware) {
	m.Handle(pattern, handler, mw...)
}

// Group will register the provided handler under the prefix. The prefix must
// end with a trailing slash.
func (m *Mux) Group(prefix string, h http.Handler, mw ...Middleware) {
	m.Handle(prefix, http.StripPrefix(strings.TrimSuffix(prefix, "/"), h), mw...)
}
