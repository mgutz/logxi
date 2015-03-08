package log

import (
	"os"
	"strconv"
	"strings"
)

var contextLines int

// Configuration comes from environment or external services like
// consul, etcd.
type Configuration struct {
	Format string `json:"format"`
	Colors string `json:"colors"`
	Levels string `json:"levels"`
}

func readFromEnviron() *Configuration {
	conf := &Configuration{}

	var envOrDefault = func(name, val string) string {
		result := os.Getenv(name)
		if result == "" {
			result = val
		}
		return result
	}

	conf.Levels = envOrDefault("LOGXI", defaultLogxiEnv)
	conf.Format = envOrDefault("LOGXI_FORMAT", defaultFormat)
	conf.Colors = envOrDefault("LOGXI_COLORS", defaultScheme)
	return conf
}

// ProcessEnv (re)processes environment.
func ProcessEnv(env *Configuration) {
	// TODO: allow reading from etcd

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
		case "fit":
			isPretty = !(value == "true" || value == "1" || value == "")
		case "maxcol":
			col, err := strconv.Atoi(value)
			if err == nil {
				maxCol = col
			} else {
				maxCol = defaultMaxCol
			}
		case "context":
			lines, err := strconv.Atoi(value)
			if err == nil {
				contextLines = lines
			} else {
				contextLines = defaultContextLines
			}
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
		// * => defaults to DBG. If someone took the time to
		// enable it ad-hoc, it probably means they are debugging
		if strings.HasPrefix(key, "-") {
			logxiNameLevelMap[key[1:]] = LevelOff
			delete(logxiNameLevelMap, key)
			key = key[1:]
			continue
		} else if value == "" {
			logxiNameLevelMap[key] = LevelAll
			continue
		}

		level := LevelAtoi[value]
		if level == 0 {
			InternalLog.Error("Unknown level in LOGXI environment variable", "key", key, "value", value, "LOGXI", env.Levels)
			level = defaultLevel
		}
		logxiNameLevelMap[key] = level
	}

	// must always have global default, otherwise errs may get eaten up
	if _, ok := logxiNameLevelMap["*"]; !ok {
		logxiNameLevelMap["*"] = LevelError
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
		colors = defaultScheme
	}
	theme = parseTheme(colors)
}
