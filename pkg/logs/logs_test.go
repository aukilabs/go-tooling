package logs

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToString(t *testing.T) {
	utests := []struct {
		in  interface{}
		out string
	}{
		{
			in:  "hello",
			out: "hello",
		},
		{
			in:  []byte("bye"),
			out: "bye",
		},
		{
			in:  -42,
			out: "-42",
		},
		{
			in:  int64(-42),
			out: "-42",
		},
		{
			in:  int32(-42),
			out: "-42",
		},
		{
			in:  int16(-42),
			out: "-42",
		},
		{
			in:  int8(-42),
			out: "-42",
		},
		{
			in:  uint(84),
			out: "84",
		},
		{
			in:  uint64(84),
			out: "84",
		},
		{
			in:  uint32(84),
			out: "84",
		},
		{
			in:  uint16(84),
			out: "84",
		},
		{
			in:  uint8(84),
			out: "84",
		},
		{
			in:  42.42,
			out: "42.42",
		},
		{
			in:  float32(42.42),
			out: "42.42",
		},
		{
			in:  true,
			out: "true",
		},
		{
			in:  false,
			out: "false",
		},
		{
			in:  time.Minute,
			out: "1m0s",
		},
		{
			in:  map[string]string{"foo": "bar"},
			out: `{"foo":"bar"}`,
		},
	}

	for _, u := range utests {
		t.Run(reflect.TypeOf(u.in).String(), func(t *testing.T) {
			require.Equal(t, u.out, toString(u.in))
		})
	}
}
