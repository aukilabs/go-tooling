package logs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
)

var (
	// The function to encode log entries and their tags.
	Encoder func(any) ([]byte, error)
)

// Set the function that logs an entry.
func SetLogger(l func(Entry)) {
	logger = l
	SetLevel(currentLevel)
}

// Available log levels.
const (
	DebugLevel Level = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

// A log level.
type Level int

// Parses a level from a string.
func ParseLevel(v string) Level {
	switch v {
	case "debug":
		return DebugLevel

	case "info":
		return InfoLevel

	case "warning":
		return WarningLevel

	case "error":
		return ErrorLevel

	default:
		l, _ := strconv.Atoi(v)
		return Level(l)
	}
}

// Converts a level to a string.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"

	case InfoLevel:
		return "info"

	case WarningLevel:
		return "warning"

	case ErrorLevel:
		return "error"

	default:
		return strconv.Itoa(int(l))
	}
}

// Sets what log levels are logged. Levels under the given level are ignored.
func SetLevel(v Level) {
	currentLevel = v

	for i := DebugLevel; i <= ErrorLevel; i++ {
		if i < v {
			loggers[i] = empytLogger
		} else {
			loggers[i] = logger
		}
	}
}

// SetInlineEncoder is a helper function that set the error encoder to
// json.Marshal.
func SetInlineEncoder() {
	Encoder = json.Marshal
}

// SetIndentEncoder is a helper function that set the error encoder to a
// function that uses json.MarshalIndent.
func SetIndentEncoder() {
	Encoder = func(v any) ([]byte, error) {
		return json.MarshalIndent(v, "", "  ")
	}
}

// Creates a log entry.
func New() Entry {
	return entry{}
}

// Creates a log entry and set the given tag.
func WithTag(k string, v any) Entry {
	return New().WithTag(k, v)
}

// Logs an error.
func Error(err error) {
	New().Error(err)
}

type Entry interface {
	// Return the time when the entry was created.
	Time() time.Time

	// Returns the log level.
	Level() Level

	// Sets the tag key with the given value. The value is converted to a
	// string.
	WithTag(k string, v any) Entry

	// Returns the log tags.
	Tags() map[string]string

	// Logs the given values with debug level.
	Debug(v ...any)

	// Logs the velues with the given format on with debug level.
	Debugf(format string, v ...any)

	// Logs the given values with info level.
	Info(v ...any)

	// Logs the velues with the given format on with info level.
	Infof(format string, v ...any)

	// Logs the given values with warning level.
	Warn(v ...any)

	// Logs the velues with the given format on with warning level.
	Warnf(format string, v ...any)

	// Logs error on error level.
	Error(err error)

	// Returns the error used to create the entry.
	GetError() error

	// Return the entry as a string.
	String() string
}

var (
	loggers       = make(map[Level]func(Entry), ErrorLevel+1)
	logger        func(e Entry)
	defaultLogger = func(e Entry) { fmt.Println(e) }
	empytLogger   = func(Entry) {}
	currentLevel  Level
)

func init() {
	SetInlineEncoder()
	SetLogger(defaultLogger)
}

func log(e Entry) {
	loggers[e.Level()](e)
}

type entry struct {
	time    time.Time
	level   Level
	message string
	tags    map[string]string
	err     error
}

func (e entry) Time() time.Time {
	return e.time
}

func (e entry) Level() Level {
	return e.level
}

func (e entry) WithTag(k string, v any) Entry {
	if e.tags == nil {
		e.tags = make(map[string]string)
	}

	e.tags[k] = toString(v)
	return e
}

func (e entry) Tags() map[string]string {
	return e.tags
}

func (e entry) Debug(v ...any) {
	e.time = time.Now()
	e.level = DebugLevel
	e.message = fmt.Sprint(v...)
	log(e)
}

func (e entry) Debugf(format string, v ...any) {
	e.time = time.Now()
	e.level = DebugLevel
	e.message = fmt.Sprintf(format, v...)
	log(e)
}

func (e entry) Info(v ...any) {
	e.time = time.Now()
	e.level = InfoLevel
	e.message = fmt.Sprint(v...)
	log(e)
}

func (e entry) Infof(format string, v ...any) {
	e.time = time.Now()
	e.level = InfoLevel
	e.message = fmt.Sprintf(format, v...)
	log(e)
}

func (e entry) Warn(v ...any) {
	e.time = time.Now()
	e.level = WarningLevel
	e.message = fmt.Sprint(v...)
	log(e)
}

func (e entry) Warnf(format string, v ...any) {
	e.time = time.Now()
	e.level = WarningLevel
	e.message = fmt.Sprintf(format, v...)
	log(e)
}

func (e entry) Error(err error) {
	e.time = time.Now()
	e.level = ErrorLevel
	e.message = errors.Message(err)
	e.err = err
	log(e)
}

func (e entry) GetError() error {
	return e.err
}

func (e entry) String() string {
	b, _ := e.MarshalJSON()
	return string(b)
}

func (e entry) MarshalJSON() ([]byte, error) {
	var line string
	var typ string
	var wrappedErr error

	if err, ok := e.err.(errors.Error); ok {
		line = err.Line()
		typ = err.Type()
		wrappedErr = err.Unwrap()

		if e.tags == nil {
			e.tags = err.Tags()
		} else {
			for k, v := range err.Tags() {
				e.tags[k] = v
			}
		}
	}

	return Encoder(struct {
		Time    time.Time         `json:"time"`
		Level   string            `json:"level"`
		Message string            `json:"message"`
		Line    string            `json:"line,omitempty"`
		Type    string            `json:"type,omitempty"`
		Tags    map[string]string `json:"tags,omitempty"`
		Wrap    error             `json:"wrap,omitempty"`
	}{
		Time:    e.time,
		Level:   e.level.String(),
		Message: e.message,
		Tags:    e.tags,
		Line:    line,
		Type:    typ,
		Wrap:    wrappedErr,
	})
}

func toString(v any) string {
	switch v := v.(type) {
	case string:
		return v

	case error:
		return v.Error()

	case int:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)

	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)

	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)

	case bool:
		return strconv.FormatBool(v)

	case time.Duration:
		return v.String()

	case []byte:
		return string(v)

	default:
		b, _ := Encoder(v)
		return string(b)
	}
}
