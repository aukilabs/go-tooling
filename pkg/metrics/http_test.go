package metrics

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTP(t *testing.T) {
	s := httptest.NewServer(HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, err := io.ReadAll(r.Body)
		if err != nil || len(name) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte("Hello, "))
		w.Write(name)
	}), func(path string) string {
		return path
	}))
	defer s.Close()

	transport := HTTPTransport(http.DefaultTransport, func(path string) string {
		return path
	})

	t.Run("no payload is sent returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, s.URL, nil)
		res, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.NotNil(t, res.Body)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("payload sent returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, s.URL, bytes.NewBufferString("Ted"))
		res, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.NotNil(t, res.Body)
		require.Equal(t, http.StatusOK, res.StatusCode)

		greet, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "Hello, Ted", string(greet))
	})
}
