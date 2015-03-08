package log

const (
	// LevelEnv chooses level from LOGXI environment variable or defaults
	// to LevelInfo
	LevelEnv = iota
	// LevelAll is all levels
	LevelAll
	// LevelTrace is trace level and displays file and line in terminal
	LevelTrace
	// LevelDebug is debug level
	LevelDebug
	// LevelInfo is info level
	LevelInfo
	// LevelWarn is warn level and display file and line in terminal
	LevelWarn
	// LevelError is error level
	LevelError
	// LevelFatal is fatal level
	LevelFatal
	// LevelOff means logging is disabled for logger. This should always
	// be last.
	LevelOff
)

// FormatHappy uses HappyDevFormatter
const FormatHappy = "happy"

// FormatText uses TextFormatter
const FormatText = "text"

// FormatJSON uses JSONFormatter
const FormatJSON = "JSON"

// FormatEnv selects formatter based on LOGXI_FORMAT environment variable
const FormatEnv = ""

// LevelMap maps int enums to string level.
var LevelMap = map[int]string{
	LevelTrace: "TRC",
	LevelDebug: "DBG",
	LevelInfo:  "INF",
	LevelWarn:  "WRN",
	LevelError: "ERR",
	LevelFatal: "FTL",
}

// LevelMap maps int enums to string level.
var LevelAtoi = map[string]int{
	"ALL":   LevelAll,
	"TRC":   LevelTrace,
	"DBG":   LevelDebug,
	"INF":   LevelInfo,
	"WRN":   LevelWarn,
	"ERR":   LevelError,
	"FTL":   LevelFatal,
	"OFF":   LevelOff,
	"all":   LevelAll,
	"trace": LevelTrace,
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
	"fatal": LevelFatal,
	"off":   LevelOff,
}

// Logger is the interface for logging.
type Logger interface {
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Log(level int, msg string, args []interface{})

	SetLevel(int)
	IsTrace() bool
	IsDebug() bool
	IsInfo() bool
	IsWarn() bool
	// Error, Fatal not needed, those SHOULD always be logged
}
