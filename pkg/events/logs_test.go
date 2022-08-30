package events

import (
	"testing"
	"time"

	"github.com/aukilabs/go-tooling/pkg/logs"
)

func TestLogger(t *testing.T) {
	initLogs(t)

	l := Logger{
		Pusher: &Pusher{
			Endpoint:      "",
			BatchSize:     100,
			FlushInterval: time.Minute,
		},
		Printer: t.Logf,
	}

	logs.SetLogger(l.Log)

	logs.New().Debug("hi")
}
