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

func init() {
	// builds the format string for key value pairs up to
	// an arbitrary amount
	formatCode := Separator + "%s=%v"
	format := formatCode
	textFormatterKVFormat = map[int]string{}
	for i := 2; i < maxTextFormatterArgs; i += 2 {
		textFormatterKVFormat[i] = format
		format += formatCode
	}
}

// NewTextFormatter returns a new instance of TextFormatter. SetName
// must be called befored using it.
func NewTextFormatter(name string) *TextFormatter {
	var buildKV = func(level string) string {
		return Separator + "n=" + name + Separator + "l=" + level + Separator + "m="
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
	buf.WriteString(time.Now().Format("t=2006-01-02T15:04:05-0700"))
	buf.WriteString(tf.itoaLevelMap[level])
	buf.WriteString(msg)
	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
			if lenArgs < maxTextFormatterArgs {
				fmt.Fprintf(buf, textFormatterKVFormat[lenArgs], args...)
			} else {
				for i := 0; i < lenArgs; i += 2 {
					fmt.Fprintf(buf, "%s%s=%v", Separator, args[i], args[i+1])
				}
			}
		} else {
			buf.WriteString("IMBALANCED_PAIRS=>")
			fmt.Fprint(buf, args...)
		}
	}
	buf.WriteRune('\n')
}
