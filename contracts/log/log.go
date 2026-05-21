package log

import (
	"context"
	"fmt"
	"time"
)

// StdLogger is a standard library logger implementation.
type StdLogger struct{}

// NewStdLogger creates a new standard library logger.
func NewStdLogger() *StdLogger {
	return &StdLogger{}
}

// Debugf logs a debug message.
func (l *StdLogger) Debugf(format string, args ...any) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

// Infof logs an info message.
func (l *StdLogger) Infof(format string, args ...any) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// Warningf logs a warning message.
func (l *StdLogger) Warningf(format string, args ...any) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

// Warning logs a warning message.
func (l *StdLogger) Warning(args ...any) {
	fmt.Printf("[WARN] %v\n", args...)
}

// Errorf logs an error message.
func (l *StdLogger) Errorf(format string, args ...any) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// NoopLogger is a no-op logger that discards all messages.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Debugf does nothing.
func (l *NoopLogger) Debugf(format string, args ...any) {}

// Infof does nothing.
func (l *NoopLogger) Infof(format string, args ...any) {}

// Warningf does nothing.
func (l *NoopLogger) Warningf(format string, args ...any) {}

// Warning does nothing.
func (l *NoopLogger) Warning(args ...any) {}

// Errorf does nothing.
func (l *NoopLogger) Errorf(format string, args ...any) {}

const (
	StackDriver  = "stack"
	SingleDriver = "single"
	DailyDriver  = "daily"
	CustomDriver = "custom"
)

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
)

type Data map[string]any

type Log interface {
	// Debugf logs a message at DebugLevel.
	Debugf(format string, args ...any)
	// Infof logs a message at InfoLevel.
	Infof(format string, args ...any)
	// Warningf logs a message at WarningLevel.
	Warningf(format string, args ...any)
	// Warning logs a message at WarningLevel.
	Warning(args ...any)
	// Errorf logs a message at ErrorLevel.
	Errorf(format string, args ...any)
}

type Writer interface {
	// Debug logs a message at DebugLevel.
	Debug(args ...any)
	// Debugf is equivalent to Debug, but with support for fmt.Printf-style arguments.
	Debugf(format string, args ...any)
	// Info logs a message at InfoLevel.
	Info(args ...any)
	// Infof is equivalent to Info, but with support for fmt.Printf-style arguments.
	Infof(format string, args ...any)
	// Warning logs a message at WarningLevel.
	Warning(args ...any)
	// Warningf is equivalent to Warning, but with support for fmt.Printf-style arguments.
	Warningf(format string, args ...any)
	// Error logs a message at ErrorLevel.
	Error(args ...any)
	// Errorf is equivalent to Error, but with support for fmt.Printf-style arguments.
	Errorf(format string, args ...any)
	// Fatal logs a message at FatalLevel.
	Fatal(args ...any)
	// Fatalf is equivalent to Fatal, but with support for fmt.Printf-style arguments.
	Fatalf(format string, args ...any)
	// Panic logs a message at PanicLevel.
	Panic(args ...any)
	// Panicf is equivalent to Panic, but with support for fmt.Printf-style arguments.
	Panicf(format string, args ...any)
	// Code set a code or slug that describes the error.
	// Error messages are intended to be read by humans, but such code is expected to
	// be read by machines and even transported over different services.
	Code(code string) Writer
	// Hint set a hint for faster debugging.
	Hint(hint string) Writer
	// In sets the feature category or domain in which the log entry is relevant.
	In(domain string) Writer
	// Owner set the name/email of the colleague/team responsible for handling this error.
	// Useful for alerting purpose.
	Owner(owner any) Writer
	// Tags add multiple tags, describing the feature returning an error.
	Tags(tags ...string) Writer
	// User sets the user associated with the log entry.
	User(user any) Writer
	// With adds key-value pairs to the context of the log entry
	With(data map[string]any) Writer
	// WithTrace adds a stack trace to the log entry.
	WithTrace() Writer
}

type Logger interface {
	// Handle pass a channel config path here
	Handle(channel string) (Hook, error)
}

type Hook interface {
	// Levels monitoring level
	Levels() []Level
	// Fire executes logic when trigger
	Fire(Entry) error
}

type Entry interface {
	// Code returns the associated code.
	Code() string
	// Context returns the context of the entry.
	Context() context.Context
	// Data returns the data of the entry.
	Data() Data
	// Domain returns the domain of the entry.
	Domain() string
	// Hint returns the hint of the entry.
	Hint() string
	// Level returns the level of the entry.
	Level() Level
	// Message returns the message of the entry.
	Message() string
	// Owner returns the log's owner.
	Owner() any
	// Tags returns the list of tags.
	Tags() []string
	// Time returns the timestamp of the entry.
	Time() time.Time
	// Trace returns the stack trace or trace data.
	Trace() map[string]any
	// User returns the user information.
	User() any
	// With returns additional context data.
	With() map[string]any
}
