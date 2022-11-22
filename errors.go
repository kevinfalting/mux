package mux

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ErrorHandler holds resources for returning errors from handlers. If the writer
// is nil, it will not write the error to it. You can use the writer to capture
// a log of errors being returned to the handler. The errFunc uses http.Error if
// no function is provided.
type ErrorHandler struct {
	writer  io.Writer
	errFunc func(w http.ResponseWriter, error string, code int)
}

// NewErrorHandler creates a new ErrorHandler with the provided options.
func NewErrorHandler(options ...errOption) *ErrorHandler {
	eh := ErrorHandler{
		writer:  os.Stderr,
		errFunc: http.Error,
	}

	for _, opt := range options {
		opt(&eh)
	}

	return &eh
}

type errOption func(eh *ErrorHandler)

// WithErrWriter sets the writer on the ErrorHandler
func WithErrWriter(w io.Writer) errOption {
	return func(eh *ErrorHandler) {
		eh.writer = w
	}
}

// WithErrFunc sets the function to call when responding to the client with an error.
func WithErrFunc(f func(w http.ResponseWriter, error string, code int)) errOption {
	return func(eh *ErrorHandler) {
		eh.errFunc = f
	}
}

type errHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Err will accept a handler that can return an error and handle it according to
// the errFunc provided or http.Error by default.
func (eh *ErrorHandler) Err(h errHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var e interface{ StatusMsg() (int, string) }
		if errors.As(err, &e) {
			status, msg := e.StatusMsg()
			eh.errFunc(w, msg, status)
		} else {
			eh.errFunc(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		if eh.writer != nil {
			fmt.Fprint(eh.writer, err)
		}
	})
}

type handlerError struct {
	err    error
	status int
	msg    string
}

// StatusMsg will return the http status code and message to return to the client.
func (h *handlerError) StatusMsg() (int, string) {
	return h.status, h.msg
}

// Error satisfies the error interface
func (h *handlerError) Error() string {
	return fmt.Sprintf("status=%d msg=%q err=%q\n", h.status, h.msg, h.err)
}

// Error will return an error that can be used by the ErrorHandler
func Error(err error, status int, msg string) error {
	return &handlerError{
		err:    err,
		status: status,
		msg:    msg,
	}
}
