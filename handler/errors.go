package handler

import (
	"fmt"
	"log"
	"net/http"
)

type customHandler func(http.ResponseWriter, *http.Request) *HTTPError

type HTTPError struct {
	code      int
	internal  interface{}
	context   string
	publicMsg string
}

func (h HTTPError) Error() string {
	if h.context != "" {
		return fmt.Sprintf("HTTP %d: %s: %s", h.code, h.context, h.internal)
	}

	return fmt.Sprintf("HTTP %d: %s", h.code, h.internal)
}

func newHTTPError(code int, err interface{}) *HTTPError {
	return &HTTPError{
		code:     code,
		internal: err,
	}
}

func (e *HTTPError) Wrap(ctx string) *HTTPError {
	e.context = ctx
	return e
}

func (e *HTTPError) Code() int {
	return e.code
}

func (e *HTTPError) PublicErrMsg(msg string) *HTTPError {
	e.publicMsg = msg
	return e
}

func (h HTTPError) publicError() string {
	if h.publicMsg != "" {
		return h.publicMsg
	}

	return h.Error()
}

type ErrorHandler struct {
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (eh ErrorHandler) Wrap(handler customHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorMsg string
		err := handler(w, r)

		if err != nil {
			log.Printf("Error: %s", err.Error())

			if err.code < http.StatusInternalServerError {
				errorMsg = err.publicError()
			}

			http.Error(w, errorMsg, err.code)
		}
	})
}
