package log

import (
	"os"
	"runtime"
	"sync"
)

// DefaultLogLog is the default log for this package.
var DefaultLog Logger
var internalLog Logger

// Whether to force disabling of Colors
var disableColors bool

type loggerMap struct {
	sync.Mutex
	loggers map[string]Logger
}

var loggers = &loggerMap{
	loggers: map[string]Logger{},
}

func (lm *loggerMap) set(name string, logger Logger) {
	lm.loggers[name] = logger
}

// logxiEnabledMap maps log name patterns to levels
var logxiNameLevelMap map[string]int

// logxiFormat is the formatter kind to create
var logxiFormat string

var isTTY bool

func init() {
	stat, _ := os.Stdin.Stat()
	isTTY = (stat.Mode() & os.ModeCharDevice) != 0
	disableColors = !isTTY || runtime.GOOS == "windows"
	processEnv()
	DefaultLog = New(os.Stdout, "~")
	internalLog = New(os.Stdout, "logxi")
}

func defaultFormatterFactory(name string, kind string) (Formatter, error) {
	if kind == FormatEnv {
		kind = logxiFormat
	}

	if kind == FormatJSON {
		return NewJSONFormatter(name), nil
	}

	if disableColors {
		return NewTextFormatter(name), nil
	}

	return NewHappyDevFormatter(name), nil
}

// CreaetFormatter creates formatters and can be overriden. It accepts
// a kind in {"text", "JSON"} which correspond to TextFormatter and
// JSONFormatter, and the name of the logger.
var CreateFormatter = defaultFormatterFactory
