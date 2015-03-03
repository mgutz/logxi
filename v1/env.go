package log

import (
	"os"
	"strings"
)

// Configuration comes from environment or external services like
// consul, etcd.
type Configuration struct {
	Format string `json:"format"`
	Colors string `json:"colors"`
	Levels string `json:"levels"`
}

func readFromEnviron() *Configuration {
	conf := &Configuration{}
	conf.Levels = os.Getenv("LOGXI")
	conf.Format = os.Getenv("LOGXI_FORMAT")
	conf.Colors = os.Getenv("LOGXI_COLORS")
	return conf
}

// ProcessEnv (re)processes environment.
func ProcessEnv(env *Configuration) {
	// TODO: allow reading from etcd

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

	processLogEnv(env)
	processThemeEnv(env)
	processFormatEnv(env)
}

// processFormatEnv parses LOGXI_FORMAT
func processFormatEnv(env *Configuration) {
	logxiFormat = env.Format
	m := parseKVList(logxiFormat, ",")
	formatterFormat := ""
	tFormat := ""
	for key, value := range m {
		switch key {
		default:
			formatterFormat = key
		case "t":
			tFormat = value
		}
	}
	if formatterFormat == "" || formatterCreators[formatterFormat] == nil {
		formatterFormat = defaultFormat
	}
	logxiFormat = formatterFormat
	if tFormat == "" {
		tFormat = defaultTimeFormat
	}
	timeFormat = tFormat
}

// processLogEnv parses LOGXI variable
func processLogEnv(env *Configuration) {
	logxiEnable := env.Levels
	if logxiEnable == "" {
		logxiEnable = defaultLogxiEnv
	}

	logxiNameLevelMap = map[string]int{}
	m := parseKVList(logxiEnable, ",")
	if m == nil {
		logxiNameLevelMap["*"] = defaultLevel
	}
	for key, value := range m {
		// * => defaults to DBG because if someone took the time to
		// enable it ad-hoc, it probably means they are debugging
		if strings.HasPrefix(key, "-") {
			logxiNameLevelMap[key[1:]] = LevelOff
			delete(logxiNameLevelMap, key)
			key = key[1:]
			continue
		} else if value == "" {
			logxiNameLevelMap[key] = LevelDebug
			continue
		}

		level := LevelAtoi[value]
		if level == 0 {
			internalLog.Error("Unknown level in LOGXI environment variable", "key", key, "level", level)
			level = defaultLevel
		}
		logxiNameLevelMap[key] = level
	}

}

func getLogLevel(name string) int {
	var wildcardLevel int
	var result int

	for k, v := range logxiNameLevelMap {
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
		return LevelOff
	}

	if result > 0 {
		return result
	}

	if wildcardLevel > 0 {
		return wildcardLevel
	}

	return LevelOff
}

func processThemeEnv(env *Configuration) {
	colors := env.Colors
	if colors == "" {
		colors = DefaultScheme
	}
	theme = parseTheme(colors)
}
