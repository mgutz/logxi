package log

import (
	"fmt"
	"io"
	"runtime/debug"
	"strconv"
	"time"
)

var LTSVSeparator = "\t"

type LTSVFormatter struct {
	name         string
	itoaLevelMap map[int]string
}

func NewLTSVFormatter(name string) *LTSVFormatter {
	var buildKV = func(level string) string {
		buf := pool.Get()
		defer pool.Put(buf)

		buf.WriteString(LTSVSeparator)
		buf.WriteString("name:")
		buf.WriteString(name)

		buf.WriteString(LTSVSeparator)
		buf.WriteString("level:")
		buf.WriteString(level)

		buf.WriteString(LTSVSeparator)
		buf.WriteString("message:")

		return buf.String()
	}

	itoaLevelMap := map[int]string{
		LevelDebug: buildKV(LevelMap[LevelDebug]),
		LevelWarn:  buildKV(LevelMap[LevelWarn]),
		LevelInfo:  buildKV(LevelMap[LevelInfo]),
		LevelError: buildKV(LevelMap[LevelError]),
		LevelFatal: buildKV(LevelMap[LevelFatal]),
	}

	return &LTSVFormatter{itoaLevelMap: itoaLevelMap, name: name}
}

func (lf *LTSVFormatter) set(buf bufferWriter, key string, val interface{}) {
	buf.WriteString(LTSVSeparator)
	buf.WriteString(key)
	buf.WriteRune(':')
	if err, ok := val.(error); ok {
		buf.WriteString(err.Error())
		buf.WriteString(LTSVSeparator)

		buf.WriteString(CallStackKey)
		buf.WriteString(":")
		buf.WriteString(strconv.Quote(string(debug.Stack())))

		return
	}
	buf.WriteString(fmt.Sprintf("%v", val))
}

func (lf *LTSVFormatter) Format(writer io.Writer, level int, msg string, args []interface{}) {
	buf := pool.Get()
	defer pool.Put(buf)
	buf.WriteString("time:")
	buf.WriteString(time.Now().Format(timeFormat))
	buf.WriteString(lf.itoaLevelMap[level])
	buf.WriteString(msg)
	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs == 1 {
			lf.set(buf, singleArgKey, args[0])
		} else if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "" {
						// show key is invalid
						lf.set(buf, badKeyAtIndex(i), args[i+1])
					} else {
						lf.set(buf, key, args[i+1])
					}
				} else {
					// show key is invalid
					lf.set(buf, badKeyAtIndex(i), args[i+1])
				}
			}
		} else {
			lf.set(buf, warnImbalancedKey, args)
		}
	}
	buf.WriteRune('\n')
	buf.WriteTo(writer)
}
