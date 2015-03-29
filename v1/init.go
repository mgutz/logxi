package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// scream so user fixes it
const warnImbalancedKey = "FIX_IMBALANCED_PAIRS"
const warnImbalancedPairs = warnImbalancedKey + " => "
const singleArgKey = "_"

func badKeyAtIndex(i int) string {
	return "BAD_KEY_AT_INDEX_" + strconv.Itoa(i)
}

// DefaultLogLog is the default log for this package.
var DefaultLog Logger

// internalLog is the logger used by logxi itself
var InternalLog Logger

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

var colorableStdout io.Writer = os.Stdout
var defaultContextLines = 2
var defaultFormat string
var defaultLevel int
var defaultLogxiEnv string
var defaultLogxiFormatEnv string
var defaultMaxCol = 80
var defaultPretty = false
var defaultLogxiColorsEnv string
var defaultTimeFormat string
var disableCallstack bool
var disableCheckKeys bool
var disableColors bool
var home string
var isPretty bool
var isTerminal bool
var isWindows = runtime.GOOS == "windows"
var pkgMutex sync.Mutex
var pool = NewBufferPool()
var timeFormat string
var wd string

// logxi reserved keys

// LevelKey is the index key for level
const LevelKey = "_l"

// MessageKey is the index key for message
const MessageKey = "_m"

// NameKey is the index key for name
const NameKey = "_n"

// TimeKey is the index key for time
const TimeKey = "_t"

// CallStackKey is the indexkey for callstack
const CallStackKey = "_c"

var logxiKeys = []string{LevelKey, MessageKey, NameKey, TimeKey, CallStackKey}

func setDefaults(isTerminal bool) {
	var err error
	contextLines = defaultContextLines
	wd, err = os.Getwd()
	if err != nil {
		InternalLog.Error("Could not get working directory")
	}

	if isTerminal {
		defaultLogxiEnv = "*=WRN"
		defaultLogxiFormatEnv = "happy,fit,maxcol=80,t=15:04:05.000000,context=-1"
		defaultFormat = FormatHappy
		defaultLevel = LevelWarn
		defaultTimeFormat = "15:04:05.000000"
	} else {
		defaultLogxiEnv = "*=ERR"
		defaultLogxiFormatEnv = "JSON,t=2006-01-02T15:04:05-0700"
		defaultFormat = FormatJSON
		defaultLevel = LevelError
		defaultTimeFormat = "2006-01-02T15:04:05-0700"
		disableColors = true
	}

	if isWindows {
		home = os.Getenv("HOMEPATH")
		if os.Getenv("ConEmuANSI") == "ON" {
			defaultLogxiColorsEnv = "key=cyan+h,value,misc=blue+h,source=yellow,TRC,DBG,WRN=yellow+h,INF=green+h,ERR=red+h"
		} else {
			colorableStdout = colorable.NewColorableStdout()
			defaultLogxiColorsEnv = "ERR=red,misc=cyan,key=cyan"
		}
		// DefaultScheme is a color scheme optimized for dark background
		// but works well with light backgrounds
	} else {
		home = os.Getenv("HOME")
		term := os.Getenv("TERM")
		if term == "xterm-256color" {
			defaultLogxiColorsEnv = "key=cyan+h,value,misc=blue,source=88,TRC,DBG,WRN=yellow,INF=green+h,ERR=red+h,message=magenta+h"
		} else {
			defaultLogxiColorsEnv = "key=cyan+h,value,misc=blue,source=magenta,TRC,DBG,WRN=yellow,INF=green,ERR=red+h"
		}
	}
}

func isReservedKey(k interface{}) (bool, error) {
	key, ok := k.(string)
	if !ok {
		return false, fmt.Errorf("Key is not a string")
	}

	// check if reserved
	for _, key2 := range logxiKeys {
		if key == key2 {
			return true, nil
		}
	}
	return false, nil
}

func init() {
	// the internal logger to report errors
	if isTerminal {
		InternalLog = NewLogger3(os.Stdout, "__logxi", NewTextFormatter("__logxi"))
	} else {
		InternalLog = NewLogger3(os.Stdout, "__logxi", NewJSONFormatter("__logxi"))
	}
	InternalLog.SetLevel(LevelError)

	isTerminal = isatty.IsTerminal(os.Stdout.Fd())
	setDefaults(isTerminal)

	RegisterFormatFactory(FormatHappy, formatFactory)
	RegisterFormatFactory(FormatText, formatFactory)
	RegisterFormatFactory(FormatJSON, formatFactory)
	ProcessEnv(readFromEnviron())

	// package logger for users
	DefaultLog = New("~")
}
