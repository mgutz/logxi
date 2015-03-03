package log

import (
	"bytes"
	"io"
)

// DefaultLogger is the default logger for this package.
type DefaultLogger struct {
	writer    io.Writer
	name      string
	level     int
	formatter Formatter
}

// NewLogger creates a new default logger.
func NewLogger(writer io.Writer, name string) Logger {
	var level int
	if name != "__logxi" {
		// if err is returned, then it means the log is disabled
		level = getLogLevel(name)
		if level == LevelOff {
			return NullLog
		}
	}

	formatter, err := createFormatter(name, logxiFormat)
	if err != nil {
		panic("Could not create formatter")
	}

	log := &DefaultLogger{
		formatter: formatter,
		writer:    writer,
		name:      name,
		level:     level,
	}

	loggers.Lock()
	loggers.loggers[name] = log
	loggers.Unlock()
	return log
}

// New creates a colorable default logger.
func New(name string) Logger {
	return NewLogger(colorableStdout, name)
}

// Debug logs a debug entry.
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.Log(LevelDebug, msg, args)
	}
}

// Info logs an info entry.
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.Log(LevelInfo, msg, args)
	}
}

// Warn logs a warn entry.
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.Log(LevelWarn, msg, args)
	}
}

// Error logs an error entry.
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	l.Log(LevelError, msg, args)
}

// Fatal logs a fatal entry then panics.
func (l *DefaultLogger) Fatal(msg string, args ...interface{}) {
	l.Log(LevelFatal, msg, args)
	panic("exit due to fatal error")
}

// Log logs a leveled entry.
func (l *DefaultLogger) Log(level int, msg string, args []interface{}) {
	var buf bytes.Buffer
	l.formatter.Format(&buf, level, msg, args)
	buf.WriteTo(l.writer)
}

// IsDebug determines if this logger logs a debug statement.
func (l *DefaultLogger) IsDebug() bool {
	return l.level <= LevelDebug
}

// IsInfo determines if this logger logs an info statement.
func (l *DefaultLogger) IsInfo() bool {
	return l.level <= LevelInfo
}

// IsWarn determines if this logger logs a warning statement.
func (l *DefaultLogger) IsWarn() bool {
	return l.level <= LevelWarn
}

// SetLevel sets the level of this logger.
func (l *DefaultLogger) SetLevel(level int) {
	l.level = level
}

// SetFormatter set the formatter for this logger.
func (l *DefaultLogger) SetFormatter(formatter Formatter) {
	l.formatter = formatter
}
