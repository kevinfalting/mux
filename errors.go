package mux

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrorHandler holds resources for returning errors from handlers. If the writer
// is nil, it will not write the error to it. You can use the writer to capture
// a log of errors being returned to the handler. The errFunc uses http.Error if
// no function is provided.
type ErrorHandler struct {
	ErrWriter io.Writer
	ErrFunc   func(w http.ResponseWriter, error string, code int)
}

// ErrHandlerFunc is the function signature for handlers that return an error.
type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Err will accept a handler that can return an error and handle it according to
// the errFunc provided or http.Error by default.
func (eh *ErrorHandler) Err(h ErrHandlerFunc) http.Handler {
	if eh.ErrFunc == nil {
		eh.ErrFunc = http.Error
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var e interface{ StatusMsg() (int, string) }
		if errors.As(err, &e) {
			status, msg := e.StatusMsg()
			eh.ErrFunc(w, msg, status)
		} else {
			eh.ErrFunc(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		if eh.ErrWriter != nil {
			fmt.Fprint(eh.ErrWriter, err)
		}
	})
}

type handlerError struct {
	err         error
	status      int
	responseMsg string
}

// StatusMsg will return the http status code and message to return to the client.
func (h *handlerError) StatusMsg() (int, string) {
	return h.status, h.responseMsg
}

// Error satisfies the error interface
func (h *handlerError) Error() string {
	return fmt.Sprintf("status=%d msg=%q err=%q\n", h.status, h.responseMsg, h.err)
}

// Error will return an error that can be used by the ErrorHandler. The error
// itself is not sent back to the client, but logged instead. The status and
// optional responseMsg are both used to respond to the client.
func Error(err error, status int, responseMsg ...string) error {
	return &handlerError{
		err:         err,
		status:      status,
		responseMsg: strings.Join(responseMsg, " "),
	}
}
