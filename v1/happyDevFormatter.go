package log

import (
	"bytes"
	"fmt"
	"strings"

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
	// always use the production formatter
	jsonFormatter *JSONFormatter
}

// NewHappyDevFormatter returns a new instance of HappyDevFormatter.
// Performance isn't priority. It's more important developers see errors
// and stack.
func NewHappyDevFormatter(name string) *HappyDevFormatter {

	return &HappyDevFormatter{
		name:          name,
		jsonFormatter: NewJSONFormatter(name),
	}
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

func (hd *HappyDevFormatter) getLevelContext(level int) (context string, color string) {
	switch level {
	case LevelDebug:
		color = theme.Debug
	case LevelInfo:
		color = theme.Info
	case LevelWarn:
		c := stack.Caller(3)
		context = fmt.Sprintf("%#v", c)
		color = theme.Warn
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
			errbuf.WriteString(fmt.Sprintf("%#v", stack))
			lines++
		}
		if lines > 1 {
			errbuf.WriteRune('\n')
		}
		context = errbuf.String()
		color = theme.Error
	}

	return context, color
}

// logxi reserved keys
const levelKey = "l"
const messageKey = "m"
const nameKey = "n"
const timeKey = "t"
const atKey = "@"

var logxiKeys = []string{atKey, levelKey, messageKey, nameKey, timeKey}

func isReservedKey(k interface{}) (bool, error) {
	// check if reserved
	if key, ok := k.(string); ok {
		for _, key2 := range logxiKeys {
			if key == key2 {
				return true, nil
			}
		}
	} else {
		return false, fmt.Errorf("Key is not a string")
	}
	return false, nil
}

// Format records a log entry.
func (hd *HappyDevFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {

	// warn about reserved keys and bad keys
	for i := 0; i < len(args); i += 2 {
		isReserved, err := isReservedKey(args[i])
		if err != nil {
			// not a string
			internalLog.Error("Key is not a string.", fmt.Sprintf("args[%d]", i), fmt.Sprintf("%v", args[i]))
		} else if isReserved {
			internalLog.Fatal("Key conflicts with reserved key. Avoiding using single rune keys.", "key", args[i].(string))
		}
	}

	// delegate to production JSON formatter, this ensures
	// there will not be any surprises in production
	entry := hd.jsonFormatter.LogEntry(level, msg, args)

	// reset the column tracker used for fancy formatting
	hd.col = 0

	buf.WriteString(theme.Misc)
	hd.writeString(buf, entry[timeKey].(string))
	buf.WriteString(theme.Reset)

	context, color := hd.getLevelContext(level)

	// DBG, INF ...
	hd.set(buf, "", entry[levelKey].(string), color)
	// logger name
	hd.set(buf, "", entry[nameKey], theme.Misc)
	// message from user
	hd.set(buf, "", entry[messageKey], color)
	// WRN,ERR file, line number context
	if context != "" {
		hd.set(buf, atKey, context, color)
	}

	// print in same order as arguments. The log entry from JSONFormatter is a
	// JSON object and likely does not match the order of the arguments.
	// Preserve order so it's easier for developers to debug.
	order := []string{}
	lenArgs := len(args)
	for i := 0; i < len(args); i += 2 {
		if i+1 >= lenArgs {
			continue
		}
		if key, ok := args[i].(string); ok {
			order = append(order, key)
		} else {
			order = append(order, badKeyAtIndex(i))
		}
	}

	for _, key := range order {
		// skip logxi keys
		isReserved, err := isReservedKey(key)
		if err != nil {
			panic("key is invalid. Should never get here. " + err.Error())
		} else if isReserved {
			continue
		}
		hd.set(buf, key, entry[key], theme.Value)
	}

	buf.WriteRune('\n')
	buf.WriteString(theme.Reset)
}
