package log

import (
	"fmt"
	"os"
	"strings"
)

func processEnv() {
	processLogEnv()
	processThemeEnv()
	processFormatEnv()
}

// processFormatEnv parses LOGXI_FORMAT
func processFormatEnv() {
	logxiFormat = os.Getenv("LOGXI_FORMAT")
	allowed := "JSON text"
	if logxiFormat == "" || !strings.Contains(allowed, logxiFormat) {
		if isTTY {
			logxiFormat = FormatText
		} else {
			logxiFormat = FormatJSON
		}
	}
}

// processLogEnv parses LOGXI variable
func processLogEnv() {
	logxiEnable := os.Getenv("LOGXI")
	if logxiEnable == "" {
		if isTTY {
			logxiEnable = "*=WRN"
		} else {
			logxiEnable = "*=ERR"
		}
	}

	logxiNameLevelMap = map[string]int{}
	pairs := strings.Split(logxiEnable, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		// * => defaults to DBG because if someone took the time to
		// enable it ad-hoc, it probably means they are debugging
		if len(kv) == 1 {
			key := kv[0]
			if strings.HasPrefix(key, "-") {
				logxiNameLevelMap[key[1:]] = LevelOff
			} else {
				logxiNameLevelMap[key] = LevelDebug
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
					internalLog.Error("Unknown level in LOGXI environment variable", "level", kv[1])
					if isTTY {
						level = LevelWarn
					} else {
						level = LevelError
					}
				}
			}
			logxiNameLevelMap[key] = level
		}
	}

}

func getLogLevel(name string) (int, error) {
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
