package logs

type Level int

const (
	Info = iota
	Warning
	Error
	Debug
)

type Entry interface {
	// Sets the tag key with the given value. The value is converted to a
	// string.
	WithTag(k string, v any) Entry

	Debug(v ...any)

	Debugf(format ...any)

	Info(v ...any)

	Infof(format string, v ...any)

	Warn(v ...any)

	Warnf(format string, v ...any)

	Error(err error)
}

type entry struct {
	level   Level
	message string
	tags    map[string]string
}

// func test() {
// 	logs.New().Error(err)

// 	logs.New().
// 		WithTag("", "").
// 		WithTag("", "").
// 		Info("My error")
// }
