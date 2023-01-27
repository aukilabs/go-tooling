package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/aukilabs/go-tooling/pkg/logs"
)

// Mux is an HTTP request multiplexer. It matches the URL of each incoming
// request against request methods and a list of registered patterns and calls
// the handler for the pattern that most closely matches the URL.
type Mux struct {
	muxes map[string]*http.ServeMux
}

// NewMux creates a Mux.
func NewMux() Mux {
	return Mux{
		muxes: make(map[string]*http.ServeMux),
	}
}

// HandleFunc registers the handler function for the given method and pattern.
// The documentation for ServeMux explains how patterns are matched.
func (m Mux) HandleFunc(method string, pattern string, h http.HandlerFunc) {
	m.Handle(method, pattern, h)
}

// Handle registers the handler for the given method and pattern. The
// documentation for http.ServeMux explains how patterns are matched.
func (m Mux) Handle(method string, pattern string, h http.Handler) {
	mux, ok := m.muxes[method]
	if !ok {
		mux = http.NewServeMux()
		m.muxes[method] = mux
	}
	mux.Handle(pattern, h)
}

// ServeHTTP dispatches the request to the handler whose method and pattern most
// closely match the request URL.
func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux, ok := m.muxes[r.Method]
	if !ok {
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}
	mux.ServeHTTP(w, r)
}

// BaseHandler is a struct to embed that provides helper methods to handle a
// RESTful API.
type BaseHandler struct {
	Encoder func(any) ([]byte, error)
	Decoder func([]byte, any) error
}

// Error writes an HTTP error.
func (h BaseHandler) Error(w http.ResponseWriter, r *http.Request, code int, err error) {
	if code >= 500 {
		logs.Error(errors.New("http request failed").
			WithTag("method", r.Method).
			WithTag("path", r.URL.Path).
			WithTag("code", code).
			Wrap(err))
	} else {
		logs.WithTag("method", r.Method).
			WithTag("path", r.URL.Path).
			WithTag("code", code).
			WithTag("error", err).
			Debug("http request failed")
	}
	http.Error(w, "", code)
}

// Error writes a bad request HTTP error.
func (h BaseHandler) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	h.Error(w, r, http.StatusBadRequest, err)
}

// Error writes an unauthorized HTTP error.
func (h BaseHandler) Unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	h.Error(w, r, http.StatusUnauthorized, err)
}

// Error writes a too many requests HTTP error.
func (h BaseHandler) TooManyRequests(w http.ResponseWriter, r *http.Request, err error) {
	h.Error(w, r, http.StatusTooManyRequests, err)
}

// Error writes an internal server error HTTP error.
func (h BaseHandler) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	h.Error(w, r, http.StatusInternalServerError, err)
}

// Error writes an not found error HTTP error.
func (h BaseHandler) NotFound(w http.ResponseWriter, r *http.Request, err error) {
	h.Error(w, r, http.StatusNotFound, err)
}

// Ok writes an ok response and encode the given values. Values are encoded as
// an array when there is more than 1.
func (h BaseHandler) Ok(w http.ResponseWriter, r *http.Request, v ...any) {
	var body []byte
	var err error

	if len(v) == 1 {
		body, err = h.encode(v[0])
	} else if len(v) != 0 {
		body, err = h.encode(v)
	}
	if err != nil {
		h.InternalServerError(w, r, errors.New("encoding response failed").Wrap(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if len(body) == 0 {
		return
	}
	w.Write(body)
}

// NotModified writes a not modified response.
func (h BaseHandler) NotModified(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotModified)
}

func (h BaseHandler) encode(v any) ([]byte, error) {
	if h.Encoder == nil {
		return json.Marshal(v)
	}
	return h.Encoder(v)
}

func (h BaseHandler) decode(b []byte, r any) error {
	if h.Decoder == nil {
		return json.Unmarshal(b, r)
	}
	return h.Decoder(b, r)
}
