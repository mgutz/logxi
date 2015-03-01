package log

// NullLog is a noop logger. Think of it as /dev/null.
var NullLog = &NullLogger{}

// NullLogger is the default logger for this package.
type NullLogger struct{}

// Debug logs a debug entry.
func (l *NullLogger) Debug(msg string, args ...interface{}) {
}

// Info logs an info entry.
func (l *NullLogger) Info(msg string, args ...interface{}) {
}

// Warn logs a warn entry.
func (l *NullLogger) Warn(msg string, args ...interface{}) {
}

// Error logs an error entry.
func (l *NullLogger) Error(msg string, args ...interface{}) {
}

// Fatal logs a fatal entry then panics.
func (l *NullLogger) Fatal(msg string, args ...interface{}) {
	panic("exit due to fatal error")
}

// Log logs a leveled entry.
func (l *NullLogger) Log(level int, msg string, args []interface{}) {
}

// IsDebug determines if this logger logs a debug statement.
func (l *NullLogger) IsDebug() bool {
	return false
}

// IsInfo determines if this logger logs an info statement.
func (l *NullLogger) IsInfo() bool {
	return false
}

// IsWarn determines if this logger logs a warning statement.
func (l *NullLogger) IsWarn() bool {
	return false
}

// SetLevel sets the level of this logger.
func (l *NullLogger) SetLevel(level int) {
}

// SetFormatter set the formatter for this logger.
func (l *NullLogger) SetFormatter(formatter Formatter) {
}
