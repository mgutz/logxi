package log

import (
	"bytes"
	"fmt"
	"time"
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

var maxTextFormatterArgs = 10
var textFormatterKVFormat map[int]string
var textFormatterFormat = Separator + "%s%s=%v"

func init() {
	// caches the format string for key value pairs
	format := textFormatterFormat
	textFormatterKVFormat = map[int]string{}
	for i := 2; i < maxTextFormatterArgs; i += 2 {
		textFormatterKVFormat[i] = format
		format += textFormatterFormat
	}
}

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

// Format records a log entry.
func (tf *TextFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	buf.WriteString("t=")
	buf.WriteString(time.Now().Format(theme.TimeFormat))
	buf.WriteString(tf.itoaLevelMap[level])
	buf.WriteString(msg)
	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
			// prints up to maxTextFormattersArgs
			fmt.Fprintf(buf, textFormatterKVFormat[lenArgs], args...)
			// prints rest
			if lenArgs >= maxTextFormatterArgs {
				for i := maxTextFormatterArgs; i < lenArgs; i += 2 {
					fmt.Fprintf(buf, textFormatterFormat, args[i], args[i+1])
				}
			}
		} else {
			buf.WriteString("IMBALANCED_KV_ARGS =>")
			fmt.Fprint(buf, args...)
		}
	}
	buf.WriteRune('\n')
}
