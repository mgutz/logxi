package log

import (
	"os"
	"sync"

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
var defaultTimeFormat string

func init() {
	// the internal log must always work and not be colored
	internalLog = NewLogger(os.Stdout, "logxi")
	internalLog.SetLevel(LevelError)
	internalLog.SetFormatter(NewTextFormatter("logxi"))

	isTerminal = isatty.IsTerminal(os.Stdout.Fd())
	ProcessEnv()
	DefaultLog = New("~")
}
