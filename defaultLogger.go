package logxi

import (
	"errors"
	"io"
)

// DefaultLogger is the default logger for this package.
type DefaultLogger struct {
	writer    io.Writer
	name      string
	level     int
	formatter Formatter
}

// NewLogger creates a new default logger. If writer is not concurrent
// safe, wrap it with NewConcurrentWriter.
func NewLogger(writer io.Writer, name string) Logger {
	formatter, err := createFormatter(name, logxiFormat)
	if err != nil {
		panic("Could not create formatter")
	}
	return NewLogger3(writer, name, formatter)
}

// NewLogger3 creates a new logger with a writer, name and formatter. If writer is not concurrent
// safe, wrap it with NewConcurrentWriter.
func NewLogger3(writer io.Writer, name string, formatter Formatter) Logger {
	var level int
	if name != "__logxi" {
		// if err is returned, then it means the log is disabled
		level = getLogLevel(name)
		if level == LevelOff {
			return NullLog
		}
	}

	log := &DefaultLogger{
		formatter: formatter,
		writer:    writer,
		name:      name,
		level:     level,
	}

	// TODO loggers will be used when watching changes to configuration such
	// as in consul, etcd
	loggers.Lock()
	loggers.loggers[name] = log
	loggers.Unlock()
	return log
}

// New creates a colorable default logger.
func New(name string) Logger {
	return NewLogger(colorableStdout, name)
}

// Trace logs a debug entry.
func (l *DefaultLogger) Trace(msg string, args ...interface{}) {
	l.traceFromFrame(0, msg, args...)
}

// TraceFromFrame logs a trace entry.
func (l *DefaultLogger) traceFromFrame(start int, msg string, args ...interface{}) {
	l.log(LevelTrace, msg, args, start)
}

// Debug logs a debug entry.
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	l.Log(LevelDebug, msg, args)
}

// Info logs an info entry.
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	l.Log(LevelInfo, msg, args)
}

// Warn logs a warn entry. If any argument is an error, then Warn returns that error
// otherwise it returns nil.
func (l *DefaultLogger) Warn(msg string, args ...interface{}) error {
	return l.warnFromFrame(0, msg, args...)
}

// WarnFromFrame logs a warn entry. If any argument is an error, then Warn returns that error
// otherwise it returns nil.
func (l *DefaultLogger) warnFromFrame(start int, msg string, args ...interface{}) error {
	if l.IsWarn() {
		defer l.log(LevelWarn, msg, args, start)
	}
	// return original error if passed in
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			return err
		}
	}
	return nil
}

func (l *DefaultLogger) extractError(msg string, args []interface{}) error {
	var err error

	// return original error if passed in
forLoop:
	for _, arg := range args {
		switch t := arg.(type) {
		case error:
			err = t
			break forLoop
		}
	}
	if err == nil {
		err = errors.New(msg)
	}
	return err
}

// Error logs an error entry.
func (l *DefaultLogger) Error(msg string, args ...interface{}) error {
	return l.errorFromFrame(0, msg, args...)
}

// ErrorFromFrame logs an error entry.
func (l *DefaultLogger) errorFromFrame(start int, msg string, args ...interface{}) error {
	err := l.extractError(msg, args)
	l.log(LevelError, msg, args, start)
	return err
}

// Fatal logs a fatal entry then panics.
func (l *DefaultLogger) Fatal(msg string, args ...interface{}) {
	err := l.extractError(msg, args)
	l.log(LevelFatal, msg, args, 0)
	panic(err)
}

// Log logs a leveled entry.
func (l *DefaultLogger) Log(level int, msg string, args []interface{}) {
	l.log(level, msg, args, 0)
}

// Log logs a leveled entry.
func (l *DefaultLogger) log(level int, msg string, args []interface{}, startFrame int) {
	// log if the log level (warn=4) >= level of message (err=3)
	if l.level < level || silent {
		return
	}
	b, err := l.formatter.Format(level, msg, args, startFrame)
	if err != nil {
		InternalLog.Error("Unable to log", "level", level, "msg", msg, "args", args, "startFrame", startFrame)
		return
	}
	l.writer.Write(b)
}

// IsTrace determines if this logger logs a debug statement.
func (l *DefaultLogger) IsTrace() bool {
	// DEBUG(7) >= TRACE(10)
	return l.level >= LevelTrace
}

// IsDebug determines if this logger logs a debug statement.
func (l *DefaultLogger) IsDebug() bool {
	return l.level >= LevelDebug
}

// IsInfo determines if this logger logs an info statement.
func (l *DefaultLogger) IsInfo() bool {
	return l.level >= LevelInfo
}

// IsWarn determines if this logger logs a warning statement.
func (l *DefaultLogger) IsWarn() bool {
	return l.level >= LevelWarn
}

// SetLevel sets the level of this logger.
func (l *DefaultLogger) SetLevel(level int) {
	l.level = level
}

// SetFormatter set the formatter for this logger.
func (l *DefaultLogger) SetFormatter(formatter Formatter) {
	l.formatter = formatter
}
