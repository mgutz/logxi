package log

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/mgutz/ansi"
	"gopkg.in/stack.v1"
)

// Theme defines a color theme for HappyDevFormatter
type colorScheme struct {
	Key   string
	Value string
	Misc  string

	Debug string
	Info  string
	Warn  string
	Error string
	Reset string
}

const assignmentChar = ": "

var indent = "  "
var maxCol = defaultMaxCol
var theme *colorScheme

func parseKVList(s, separator string) map[string]string {
	pairs := strings.Split(s, separator)
	if len(pairs) == 0 {
		return nil
	}
	m := map[string]string{}
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "=")
		switch len(parts) {
		case 1:
			m[parts[0]] = ""
		case 2:
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func parseTheme(theme string) *colorScheme {
	m := parseKVList(theme, ",")
	var color = func(key string) string {
		style := m[key]
		c := ansi.ColorCode(style)
		if c == "" {
			c = ansi.ColorCode("reset")
		}
		//fmt.Printf("plain=%b [%s] %s=%q\n", ansi.Plain, key, style, c)
		return c
	}
	result := &colorScheme{
		Key:   color("key"),
		Value: color("value"),
		Misc:  color("misc"),
		Debug: color("DBG"),
		Warn:  color("WRN"),
		Info:  color("INF"),
		Error: color("ERR"),
		Reset: color("reset"),
	}
	return result
}

func keyColor(s string) string {
	return theme.Key + s + theme.Reset
}

// DisableColors disables coloring of log entries.
func DisableColors(val bool) {
	disableColors = val
}

// HappyDevFormatter is the formatter used for terminals. It is
// colorful, dev friendly and provides meaningful logs when
// warnings and errors occur. DO NOT use in production
type HappyDevFormatter struct {
	name string
	col  int
}

// NewHappyDevFormatter returns a new instance of HappyDevFormatter.
// Performance isn't priority. It's more important developers see errors
// and stack.
func NewHappyDevFormatter(name string) *HappyDevFormatter {
	return &HappyDevFormatter{name: name}
}

func (hd *HappyDevFormatter) writeKey(buf *bytes.Buffer, key string) {
	// assumes this is not the first key
	hd.writeString(buf, Separator)
	if key == "" {
		return
	}
	buf.WriteString(theme.Key)
	hd.writeString(buf, key)
	hd.writeString(buf, assignmentChar)
	buf.WriteString(theme.Reset)
}

func (hd *HappyDevFormatter) offset(buf *bytes.Buffer, color string, key string, value string) {
	val := strings.Trim(value, "\n ")

	if (isPretty && key != "") || hd.col+len(key)+1+len(val) >= maxCol {
		// 4 spc
		buf.WriteString("\n")
		hd.col = 0
		hd.writeString(buf, indent)
	}
	hd.writeKey(buf, key)
	if color != "" {
		buf.WriteString(color)
	}
	hd.writeString(buf, val)

	if color != "" {
		buf.WriteString(theme.Reset)
	}
}

func (hd *HappyDevFormatter) writeError(buf *bytes.Buffer, key string, err *errors.Error) {
	msg := err.Error()
	stack := string(err.Stack())
	hd.offset(buf, theme.Error, key, msg+"\n"+stack)
}

func (hd *HappyDevFormatter) set(buf *bytes.Buffer, key string, value interface{}, color string) {
	if err, ok := value.(error); ok {
		err2 := errors.Wrap(err, 4)
		hd.writeError(buf, key, err2)
	} else if err, ok := value.(*errors.Error); ok {
		hd.writeError(buf, key, err)
	} else {
		hd.offset(buf, color, key, fmt.Sprintf("%v", value))
	}
}

// tracks the position of the string so we can break lines cleanly.
// do not send ANSI escape sequences, just raw strings
func (hd *HappyDevFormatter) writeString(buf *bytes.Buffer, s string) {
	buf.WriteString(s)
	hd.col += len(s)
}

// Format records a log entry.
func (hd *HappyDevFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	// reset the column tracker
	hd.col = 0

	buf.WriteString(theme.Misc)
	hd.writeString(buf, time.Now().Format(timeFormat))
	buf.WriteString(theme.Reset)

	var colorCode string
	var context string

	switch level {
	case LevelDebug:
		colorCode = theme.Debug
	case LevelInfo:
		colorCode = theme.Info
	case LevelWarn:
		c := stack.Caller(3)
		context = fmt.Sprintf("%+v", c)
		colorCode = theme.Warn
	default:
		trace := stack.Trace().TrimRuntime()

		// if one line, keep it on same line, multiple lines group all
		// on next line
		var errbuf bytes.Buffer
		lines := 0
		for i, stack := range trace {
			if i < 3 {
				continue
			}
			if i == 3 && len(trace) > 4 {
				errbuf.WriteString("\n\t")
			} else if i > 3 {
				errbuf.WriteString("\n\t")
			}
			errbuf.WriteString(fmt.Sprintf("%+v", stack))
			lines++
		}
		if lines > 1 {
			errbuf.WriteRune('\n')
		}

		context = errbuf.String()
		colorCode = theme.Error
	}
	// DBG, INF ...
	hd.set(buf, "", LevelMap[level], colorCode)
	// logger name
	hd.set(buf, "", hd.name, theme.Misc)
	// message from user
	hd.set(buf, "", msg, colorCode)

	// WRN,ERR file, line number context
	if context != "" {
		hd.set(buf, "at", context, colorCode)
	}

	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					hd.set(buf, key, args[i+1], theme.Value)
				} else {
					hd.set(buf, "BADKEY_NAME_"+strconv.Itoa(i+1), args[i], theme.Error)
					hd.set(buf, "BADKEY_VALUE_"+strconv.Itoa(i+1), args[i+1], theme.Error)
				}
			}
		} else {
			buf.WriteString(Separator)
			buf.WriteString(theme.Error)
			buf.WriteString(warnImbalancedPairs)
			buf.WriteString(theme.Value)
			fmt.Fprint(buf, args...)
			buf.WriteString(theme.Reset)
		}
	}
	buf.WriteRune('\n')
	buf.WriteString(theme.Reset)
}
