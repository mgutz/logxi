package log

import (
	"os"
	"runtime"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// scream so user fixes it
const warnImbalancedKey = "FIX_IMBALANCED_PAIRS"
const warnImbalancedPairs = warnImbalancedKey + " => "

// DefaultLogLog is the default log for this package.
var DefaultLog Logger

// internalLog is the logger used by logxi itself
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

var isTerminal bool
var defaultFormat string
var defaultLevel int
var defaultLogxiEnv string
var defaultScheme string
var defaultTimeFormat string
var timeFormat string
var colorableStdout = colorable.NewColorableStdout()

func setDefaults(isTerminal bool) {
	if isTerminal {
		defaultLogxiEnv = "*=WRN"
		defaultFormat = FormatHappy
		defaultLevel = LevelWarn
		defaultTimeFormat = "15:04:05.000000"
	} else {
		defaultLogxiEnv = "*=ERR"
		defaultFormat = FormatText
		defaultLevel = LevelError
		defaultTimeFormat = "2006-01-02T15:04:05-0700"
		disableColors = true
	}
	if runtime.GOOS == "windows" {
		// DefaultScheme is a color scheme optimized for dark background
		// but works well with light backgrounds
		defaultScheme = "key=cyan,value,misc=blue,DBG,WRN=yellow,INF=green,ERR=red"
	} else {
		defaultScheme = "key=cyan+h,value,misc=blue,DBG,WRN=yellow+h,INF=green+h,ERR=red+h"
	}
}

func init() {
	isTerminal = isatty.IsTerminal(os.Stdout.Fd())
	setDefaults(isTerminal)

	RegisterFormatFactory(FormatHappy, formatFactory)
	RegisterFormatFactory(FormatText, formatFactory)
	RegisterFormatFactory(FormatJSON, formatFactory)
	ProcessEnv(readFromEnviron())
	// the internal log must always work and not be colored
	internalLog = NewLogger(os.Stdout, "__logxi")
	internalLog.SetLevel(LevelError)
	internalLog.SetFormatter(NewTextFormatter("__logxi"))
	DefaultLog = New("~")
}
