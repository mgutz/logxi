package log

// Debug records a debug statement.
func Debug(msg string, args ...interface{}) {
	DefaultLog.Debug(msg, args)
}

// Info records an info statement.
func Info(msg string, args ...interface{}) {
	DefaultLog.Info(msg, args)
}

// Warn records a warning statement.
func Warn(msg string, args ...interface{}) {
	DefaultLog.Warn(msg, args)
}

// Error records an error statement.
func Error(msg string, args ...interface{}) {
	DefaultLog.Error(msg, args)
}

// Fatal records a fatal statement.
func Fatal(msg string, args ...interface{}) {
	DefaultLog.Fatal(msg, args)
}

// IsDebug determines if this logger logs a debug statement.
func IsDebug() bool {
	return DefaultLog.IsDebug()
}

// IsInfo determines if this logger logs an info statement.
func IsInfo() bool {
	return DefaultLog.IsInfo()
}

// IsWarn determines if this logger logs a warning statement.
func IsWarn() bool {
	return DefaultLog.IsWarn()
}
