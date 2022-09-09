package events

import (
	"fmt"
	"runtime"

	"github.com/aukilabs/go-tooling/pkg/logs"
)

// Tags that are mapped to a log event.
const (
	AppKeyTag        = "app-key"
	SpaceIDTag       = "space-id"
	ParticipantIDTag = "participant-id"
)

// A logger that logs on the console and generate log events.
type Logger struct {
	// The pusher to send events.
	Pusher *Pusher

	// The type of the SDK.
	SDKType string

	// The SDK version family.
	SDKVersionFamily string

	// The printer to print on console. Default is fmt.Println.
	Printer func(format string, v ...any)
}

func (l Logger) Log(e logs.Entry) {
	print := func(format string, v ...any) {
		fmt.Printf(format+"\n", v...)
	}
	if l.Printer != nil {
		print = l.Printer
	}
	print("%s", e)

	l.Pusher.NewEvent(logEvent{
		AppKey:         e.Tags()[AppKeyTag],
		AukiSDKType:    l.SDKType,
		AukiSDKVersion: l.SDKVersionFamily,
		Data: logEventData{
			Message: e.String(),
			LogType: e.Level().String(),
		},
		DeviceOS:      runtime.GOOS,
		DeviceType:    runtime.GOARCH,
		ParticipantID: e.Tags()[ParticipantIDTag],
		SpaceID:       e.Tags()[SpaceIDTag],
		Event:         "log",
		Timestamp:     e.Time().UnixMilli(),
	})
}

type logEvent struct {
	AppKey         any    `json:"app_key,omitempty"`
	AppID          string `json:"application_identifier,omitempty"`
	AppProductName string `json:"application_product_name,omitempty"`
	AppVersion     string `json:"application_version,omitempty"`
	AukiSDKBuild   string `json:"auki_sdk_build,omitempty"`
	AukiSDKType    string `json:"auki_sdk_type,omitempty"`
	AukiSDKVersion string `json:"auki_sdk_version,omitempty"`
	Data           any    `json:"data,omitempty"`
	DeviceModel    string `json:"device_model,omitempty"`
	DeviceOS       string `json:"device_operating_system,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
	Event          string `json:"event,omitempty"`
	ID             string `json:"id,omitempty"`
	ParticipantID  any    `json:"participant_id,omitempty"`
	SpaceID        any    `json:"space_id,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty"`
}

type logEventData struct {
	Message    string `json:"message,omitempty"`
	LogType    string `json:"log_type,omitempty"`
	Stacktrace string `json:"stacktrace,omitempty"`
}
