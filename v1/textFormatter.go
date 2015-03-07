package log

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-errors/errors"
)

// Separator is the separator to use between key value pairs
//var Separator = "{~}"
var Separator = " "

// Formatter records log entries.
type Formatter interface {
	Format(buf *bytes.Buffer, level int, msg string, args []interface{})
}

// TextFormatter is the default recorder used if one is unspecified when
// creating a new Logger.
type TextFormatter struct {
	name         string
	itoaLevelMap map[int]string
}

// var maxTextFormatterArgs = 10
// var textFormatterKVFormat map[int]string
// var textFormatterFormat = Separator + "%s=%v"

// func init() {
// 	// caches the format string for key value pairs
// 	format := textFormatterFormat
// 	textFormatterKVFormat = map[int]string{}
// 	for i := 2; i < maxTextFormatterArgs; i += 2 {
// 		textFormatterKVFormat[i] = format
// 		format += textFormatterFormat
// 	}
// }

// NewTextFormatter returns a new instance of TextFormatter. SetName
// must be called befored using it.
func NewTextFormatter(name string) *TextFormatter {
	var buildKV = func(level string) string {
		var buf bytes.Buffer
		buf.WriteString(Separator)
		buf.WriteString("n=")
		buf.WriteString(name)

		buf.WriteString(Separator)
		buf.WriteString("l=")
		buf.WriteString(level)

		buf.WriteString(Separator)
		buf.WriteString("m=")

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

func (tf *TextFormatter) set(buf *bytes.Buffer, key string, val interface{}) {
	buf.WriteString(Separator)
	buf.WriteString(key)
	buf.WriteRune('=')
	if err, ok := val.(error); ok {
		err2 := errors.Wrap(err, 4)
		msg := err2.Error()
		stack := string(err2.Stack())
		buf.WriteString(msg)
		buf.WriteRune('\n')
		buf.WriteString(stack)
		return

	}
	buf.WriteString(fmt.Sprintf("%#v", val))
}

// Format records a log entry.
func (tf *TextFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	buf.WriteString("t=")
	buf.WriteString(time.Now().Format(timeFormat))
	buf.WriteString(tf.itoaLevelMap[level])
	buf.WriteString(msg)
	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
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
}
