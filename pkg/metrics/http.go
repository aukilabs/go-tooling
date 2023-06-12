package metrics

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	methodLabel    = "method"
	pathLabel      = "path"
	endpointLabel  = "endpoint"
	statusLabel    = "status"
	errorTypeLabel = "error_type"
)

var (
	inboundHTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "inbound_http_requests",
		Help: "The number of inbound http requests.",
	}, []string{
		methodLabel,
		pathLabel,
		statusLabel,
	})

	inboundHTTPRequestLatencies = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "inbound_http_request_latencies",
		Help: "The time to process inbound http requests.",
	}, []string{
		methodLabel,
		pathLabel,
		statusLabel,
	})

	inboundHTTPRequestReceivedBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "inbound_http_request_received_bytes",
		Help: "The number of bytes received from inbound HTTP requests.",
	}, []string{
		methodLabel,
		pathLabel,
		errorTypeLabel,
	})

	inboundHTTPRequestSentBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "inbound_http_request_sent_bytes",
		Help: "The number of bytes sent in inbound HTTP requests.",
	}, []string{
		methodLabel,
		pathLabel,
		errorTypeLabel,
	})

	outboundHTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbound_http_requests",
		Help: "The number of outbound http requests.",
	}, []string{
		methodLabel,
		endpointLabel,
		pathLabel,
		statusLabel,
		errorTypeLabel,
	})

	outboundHTTPRequestLatencies = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "outbound_http_request_latencies",
		Help: "The time to process outbound http requests.",
	}, []string{
		endpointLabel,
		methodLabel,
		pathLabel,
		statusLabel,
		errorTypeLabel,
	})

	outboundHTTPRequestSentBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbound_http_request_sent_bytes",
		Help: "The number of bytes sent in outbound HTTP requests.",
	}, []string{
		endpointLabel,
		methodLabel,
		pathLabel,
		errorTypeLabel,
	})

	outboundHTTPRequestReceivedBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbound_http_request_received_bytes",
		Help: "The number of bytes received in outbound HTTP requests.",
	}, []string{
		endpointLabel,
		methodLabel,
		pathLabel,
		errorTypeLabel,
	})
)

// A function that formats a path.
//
// When dealing with metrics, each different path adds a dimension that has a
// toll on metrics size and aggregation performances.
//
// This is to prevent paths like the ones which include identifiers to over
// create metrics dimensions.
type PathFormater func(*http.Request, string) string

// The default path formater used when none is specified.
//
// The formater returns the first element of the path suffixed with a / when
// there are multiple elements.
func DefaultPathFormater(_ *http.Request, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	p := strings.TrimPrefix(path, "/")
	idx := strings.Index(p, "/")
	if idx < 0 || len(p) == 0 {
		return path
	}
	return path[:idx+2]
}

// Returns an HTTP handler that generates metrics for the given handler.
func HTTPHandler(h http.Handler, pathFormater ...PathFormater) http.Handler {
	if len(pathFormater) == 0 {
		pathFormater = append(pathFormater, DefaultPathFormater)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		path := r.URL.Path
		for _, f := range pathFormater {
			path = f(r, path)
		}

		if r.Body != nil {
			r.Body = makeReadCloser(r.Body, func(bytes int, err error) {
				inboundHTTPRequestReceivedBytes.With(prometheus.Labels{
					methodLabel:    r.Method,
					pathLabel:      path,
					errorTypeLabel: errors.Type(err),
				}).Add(float64(bytes))
			})
		}

		rw := makeResponseWriter(w, func(statusCode, bytes int, err error) {
			inboundHTTPRequestSentBytes.With(prometheus.Labels{
				methodLabel:    r.Method,
				pathLabel:      path,
				errorTypeLabel: errors.Type(err),
			}).Add(float64(bytes))
		})

		statusCode := strconv.Itoa(rw.statusCode)

		h.ServeHTTP(&rw, r)

		inboundHTTPRequests.With(prometheus.Labels{
			methodLabel: r.Method,
			pathLabel:   path,
			statusLabel: statusCode,
		}).Inc()

		inboundHTTPRequestLatencies.With(prometheus.Labels{
			methodLabel: r.Method,
			pathLabel:   r.URL.Path,
			statusLabel: statusCode,
		}).Observe(time.Since(start).Seconds())
	})
}

// Return an HTTP transport that generates metrics for the given transport.
func HTTPTransport(t http.RoundTripper, pathFormater ...PathFormater) http.RoundTripper {
	if len(pathFormater) == 0 {
		pathFormater = append(pathFormater, DefaultPathFormater)
	}

	return transport{
		RoundTripper:  t,
		pathFormaters: pathFormater,
	}
}

type responseWriter struct {
	http.ResponseWriter

	observe    func(statusCode int, bytes int, err error)
	statusCode int
}

func makeResponseWriter(w http.ResponseWriter, observe func(statusCode, bytes int, err error)) responseWriter {
	return responseWriter{
		ResponseWriter: w,
		observe:        observe,
		statusCode:     http.StatusOK,
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.observe(w.statusCode, n, err)
	return n, err
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack is not supported").WithType("http-hijack-not-supported")
	}
	return hj.Hijack()
}

type readCloser struct {
	io.ReadCloser

	observe func(bytes int, err error)
}

func makeReadCloser(r io.ReadCloser, observe func(bytes int, err error)) readCloser {
	return readCloser{
		ReadCloser: r,
		observe:    observe,
	}
}

func (r readCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.observe(n, err)
	return n, err
}

type transport struct {
	http.RoundTripper
	pathFormaters []PathFormater
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	path := req.URL.Path
	for _, f := range t.pathFormaters {
		path = f(req, path)
	}

	if req.Body != nil {
		req.Body = makeReadCloser(req.Body, func(bytes int, err error) {
			outboundHTTPRequestSentBytes.With(prometheus.Labels{
				methodLabel:    req.Method,
				endpointLabel:  req.URL.Host,
				pathLabel:      path,
				errorTypeLabel: errors.Type(err),
			}).Add(float64(bytes))
		})
	}

	res, err := t.RoundTripper.RoundTrip(req)
	if err == nil && res.Body != nil {
		res.Body = makeReadCloser(res.Body, func(bytes int, err error) {
			outboundHTTPRequestReceivedBytes.With(prometheus.Labels{
				methodLabel:    req.Method,
				endpointLabel:  req.URL.Host,
				pathLabel:      path,
				errorTypeLabel: errors.Type(err),
			}).Add(float64(bytes))
		})
	}

	var statusCode string
	if res != nil {
		statusCode = strconv.Itoa(res.StatusCode)
	}

	outboundHTTPRequests.With(prometheus.Labels{
		endpointLabel:  req.URL.Host,
		methodLabel:    req.Method,
		pathLabel:      path,
		statusLabel:    statusCode,
		errorTypeLabel: errors.Type(err),
	}).Inc()

	outboundHTTPRequestLatencies.With(prometheus.Labels{
		endpointLabel:  req.URL.Host,
		methodLabel:    req.Method,
		pathLabel:      path,
		statusLabel:    statusCode,
		errorTypeLabel: errors.Type(err),
	}).Observe(time.Since(start).Seconds())

	return res, err
}
