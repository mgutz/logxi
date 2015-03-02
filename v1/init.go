package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/mgutz/str"
)

// DefaultLogLog is the default log for this package.
var DefaultLog Logger

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

// logxiEnabledMap maps name patterns to levels
var logxiEnabledMap map[string]int

// logxiFormat is the formatter kind to create
var logxiFormat string

func processEnv() {
	logxiEnable := os.Getenv("LOGXI")
	if logxiEnable == "" {
		if isTTY {
			logxiEnable = "*=WRN"
		} else {
			logxiEnable = "*=ERR"
		}
	}

	logxiEnable = str.Clean(logxiEnable)
	logxiEnabledMap = map[string]int{}
	pairs := strings.Split(logxiEnable, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		// * => defaults to DBG because if someone took the time to
		// enable it ad-hoc, it probably means they are debugging
		if len(kv) == 1 {
			key := kv[0]
			if strings.HasPrefix(key, "-") {
				logxiEnabledMap[key[1:]] = LevelOff
			} else {
				logxiEnabledMap[key] = LevelDebug
			}
		} else if len(kv) == 2 {
			key := kv[0]
			level := 0
			if strings.HasPrefix(key, "-") {
				key = key[1:]
				level = LevelOff
			} else {
				level = LevelAtoi[kv[1]]
				if level == 0 {
					if isTTY {
						level = LevelWarn
					} else {
						level = LevelError
					}
				}
			}
			logxiEnabledMap[key] = level
		}
	}

	logxiFormat = os.Getenv("LOGXI_FORMAT")
	allowed := "JSON text"
	if logxiFormat == "" || !strings.Contains(allowed, logxiFormat) {
		if isTTY {
			logxiFormat = FormatText
		} else {
			logxiEnable = FormatJSON
		}
	}
}

func getLogLevel(name string) (int, error) {
	var wildcardLevel int
	var result int

	for k, v := range logxiEnabledMap {
		if k == name {
			result = v
		} else if k == "*" {
			wildcardLevel = v
		} else if strings.HasPrefix(k, "*") && strings.HasSuffix(name, k[1:]) {
			result = v
		} else if strings.HasSuffix(k, "*") && strings.HasPrefix(name, k[:len(k)-1]) {
			result = v
		}
	}

	if result == LevelOff {
		return LevelOff, fmt.Errorf("is not enabled")
	}

	if result > 0 {
		return result, nil
	}

	if wildcardLevel > 0 {
		return wildcardLevel, nil
	}

	return LevelOff, fmt.Errorf("is not enabled")
}

var isTTY bool

func init() {
	stat, _ := os.Stdin.Stat()
	isTTY = (stat.Mode() & os.ModeCharDevice) != 0
	disableColors = !isTTY || runtime.GOOS == "windows"
	processEnv()
	DefaultLog = New(os.Stdout, "")
}

func defaultFormatterFactory(name string, kind string) (Formatter, error) {
	if kind == FormatEnv {
		kind = logxiFormat
	}

	if kind == FormatJSON {
		return NewJSONFormatter(name), nil
	}

	if !disableColors {
		return NewHappyDevFormatter(name), nil
	}
	return NewTextFormatter(name), nil
}

// CreaetFormatter creates formatters and can be overriden. It accepts
// a kind in {"text", "JSON"} which correspond to TextFormatter and
// JSONFormatter, and the name of the logger.
var CreateFormatter = defaultFormatterFactory
