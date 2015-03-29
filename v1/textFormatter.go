package log

import (
	"fmt"
	"io"
	"runtime/debug"
	"time"
)

// Separator is the separator to use between key value pairs
//var Separator = "{~}"
var Separator = " "

// Formatter records log entries.
type Formatter interface {
	Format(writer io.Writer, level int, msg string, args []interface{})
}

// TextFormatter is the default recorder used if one is unspecified when
// creating a new Logger.
type TextFormatter struct {
	name         string
	itoaLevelMap map[int]string
}

// NewTextFormatter returns a new instance of TextFormatter. SetName
// must be called befored using it.
func NewTextFormatter(name string) *TextFormatter {
	var buildKV = func(level string) string {
		buf := pool.Get()
		defer pool.Put(buf)
		buf.WriteString(Separator)
		buf.WriteString("_n=")
		buf.WriteString(name)

		buf.WriteString(Separator)
		buf.WriteString("_l=")
		buf.WriteString(level)

		buf.WriteString(Separator)
		buf.WriteString("_m=")

		return buf.String()
	}
	itoaLevelMap := map[int]string{
		LevelDebug: buildKV(LevelMap[LevelDebug]),
		LevelWarn:  buildKV(LevelMap[LevelWarn]),
		LevelInfo:  buildKV(LevelMap[LevelInfo]),
		LevelError: buildKV(LevelMap[LevelError]),
		LevelFatal: buildKV(LevelMap[LevelFatal]),
	}
	return &TextFormatter{itoaLevelMap: itoaLevelMap, name: name}
}

func (tf *TextFormatter) set(buf bufferWriter, key string, val interface{}) {
	buf.WriteString(Separator)
	buf.WriteString(key)
	buf.WriteRune('=')
	if err, ok := val.(error); ok {
		buf.WriteString(err.Error())
		buf.WriteRune('\n')
		buf.WriteString(string(debug.Stack()))
		return
	}
	buf.WriteString(fmt.Sprintf("%v", val))
}

// Format records a log entry.
func (tf *TextFormatter) Format(writer io.Writer, level int, msg string, args []interface{}) {
	buf := pool.Get()
	defer pool.Put(buf)
	buf.WriteString("_t=")
	buf.WriteString(time.Now().Format(timeFormat))
	buf.WriteString(tf.itoaLevelMap[level])
	buf.WriteString(msg)
	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs == 1 {
			tf.set(buf, singleArgKey, args[0])
		} else if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "" {
						// show key is invalid
						tf.set(buf, badKeyAtIndex(i), args[i+1])
					} else {
						tf.set(buf, key, args[i+1])
					}
				} else {
					// show key is invalid
					tf.set(buf, badKeyAtIndex(i), args[i+1])
				}
			}
		} else {
			tf.set(buf, warnImbalancedKey, args)
		}
	}
	buf.WriteRune('\n')
	buf.WriteTo(writer)
}
