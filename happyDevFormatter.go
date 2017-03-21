package logxi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mgutz/ansi"
)

// colorScheme defines a color theme for HappyDevFormatter
type colorScheme struct {
	Key     string
	Message string
	Value   string
	Misc    string
	Source  string

	Trace string
	Debug string
	Info  string
	Warn  string
	Error string
}

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
	cs := &colorScheme{}
	var wildcard string

	var color = func(key string) string {
		if disableColors {
			return ""
		}
		style := m[key]
		c := ansi.ColorCode(style)
		if c == "" {
			c = wildcard
		}
		return c
	}
	wildcard = color("*")

	if wildcard != ansi.Reset {
		cs.Key = wildcard
		cs.Value = wildcard
		cs.Misc = wildcard
		cs.Source = wildcard
		cs.Message = wildcard

		cs.Trace = wildcard
		cs.Debug = wildcard
		cs.Warn = wildcard
		cs.Info = wildcard
		cs.Error = wildcard
	}

	cs.Key = color("key")
	cs.Value = color("value")
	cs.Misc = color("misc")
	cs.Source = color("source")
	cs.Message = color("message")

	cs.Trace = color("TRC")
	cs.Debug = color("DBG")
	cs.Warn = color("WRN")
	cs.Info = color("INF")
	cs.Error = color("ERR")
	return cs
}

// HappyDevFormatter is the formatter used for terminals. It is
// colorful, dev friendly and provides meaningful logs when
// warnings and errors occur.
//
// HappyDevFormatter does not worry about performance. It's at least 3-4X
// slower than JSONFormatter since it delegates to JSONFormatter to marshal
// then unmarshal JSON. Then it does other stuff like read source files, sort
// keys all to give a developer more information.
//
// SHOULD NOT be used in production for extended period of time. However, it
// works fine in SSH terminals and binary deployments.
type HappyDevFormatter struct {
	// always use the production formatter
	jsonFormatter *JSONFormatter
}

// NewHappyDevFormatter returns a new instance of HappyDevFormatter.
func NewHappyDevFormatter(name string) *HappyDevFormatter {
	jf := NewJSONFormatter(name)
	return &HappyDevFormatter{
		jsonFormatter: jf,
	}
}

func (hd *HappyDevFormatter) writeKey(buf bufferWriter, key string, col int) int {
	// assumes this is not the first key
	col = hd.writeString(buf, Separator, col)
	if key == "" {
		return col
	}
	writeColor(buf, theme.Key)
	col = hd.writeString(buf, key, col)
	col = hd.writeString(buf, AssignmentChar, col)
	writeColor(buf, ansi.Reset)
	return col
}

func (hd *HappyDevFormatter) set(buf bufferWriter, key string, value interface{}, color string, col int) int {
	var str string
	if s, ok := value.(string); ok {
		str = s
	} else if s, ok := value.(fmt.Stringer); ok {
		str = s.String()
	} else {
		str = fmt.Sprintf("%v", value)
	}
	val := strings.Trim(str, "\n ")
	if (isPretty && key != "") || col+len(key)+2+len(val) >= maxCol {
		buf.WriteString("\n")
		col = 0
		col = hd.writeString(buf, indent, col)
	}
	col = hd.writeKey(buf, key, col)
	col = hd.writeColoredString(buf, val, color, col)
	return col
}

// Write a string and tracks the position of the string so we can break lines
// cleanly. Do not send ANSI escape sequences, just raw strings
func (hd *HappyDevFormatter) writeString(buf bufferWriter, s string, col int) int {
	buf.WriteString(s)
	return col + len(s)
}

func (hd *HappyDevFormatter) writeColoredString(buf bufferWriter, s string, color string, col int) int {
	writeColor(buf, color)
	buf.WriteString(s)
	writeColor(buf, ansi.Reset)
	return col + len(s)
}

func (hd *HappyDevFormatter) sourceContext(color string, frames []Frame) string {
	if disableCallstack {
		return ""
	}

	buf := pool.Get()
	defer pool.Put(buf)

	// ignored runtime
	//frames := parseDebugStack(string(debug.Stack()), 5, true)
	if len(frames) == 0 {
		return ""
	}
	for _, frame := range frames {
		context := frame.String(color, theme.Source)
		if context != "" {
			buf.WriteString(context)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func (hd *HappyDevFormatter) levelSourceContext(level int, entry map[string]interface{}, args []interface{}, startFrame int) (context string, color string) {
	const skipLogxiFrames = 0

	getStack := func(offset int, size int) []Frame {
		// the offset skips logxi frames and is empirically determined

		if startFrame == -1 {
			// get everything
			return callstack(offset, -1, true)
		}

		result := callstack(offset+startFrame, size, true)
		return result
	}

	switch level {
	default:
		color = ""
		context = ""
	case LevelInfo:
		color = theme.Info
	case LevelWarn:
		color = theme.Warn
		kv := entry[KeyMap.CallStack]
		// no callstack means an error was not included in warning, grab 1 frame
		if kv == nil {
			context = hd.sourceContext(color, getStack(skipLogxiFrames, 1))
			break
		}
		context = hd.sourceContext(color, getStack(skipLogxiFrames, -1))
	case LevelError, LevelFatal:
		color = theme.Error
		context = hd.sourceContext(color, getStack(skipLogxiFrames, -1))

	case LevelTrace:
		color = theme.Trace
		context = hd.sourceContext(color, getStack(skipLogxiFrames, 1))
		context += "\n"
	case LevelDebug:
		color = theme.Debug
	}

	return context, color
}

// Format a log entry.
func (hd *HappyDevFormatter) Format(level int, msg string, args []interface{}, startFrame int) ([]byte, error) {
	buf := pool.Get()
	defer pool.Put(buf)

	if len(args) == 1 {
		args = append(args, 0)
		copy(args[1:], args[0:])
		args[0] = singleArgKey
	}

	// warn about reserved, bad and complex keys
	for i := 0; i < len(args); i += 2 {
		isReserved, err := isReservedKey(args[i])
		if err != nil {
			InternalLog.Error("Key is not a string.", "err", fmt.Errorf("args[%d]=%v", i, args[i]))
		} else if isReserved {
			InternalLog.Fatal("Key conflicts with reserved key. Avoiding using single rune keys.", "key", args[i].(string))
		} else {
			// Ensure keys are simple strings. The JSONFormatter doesn't escape
			// keys as a performance tradeoff. This panics if the JSON key
			// value has a different value than a simple quoted string.
			key := args[i].(string)
			b, err := json.Marshal(key)
			if err != nil {
				panic("Key is invalid. " + err.Error())
			}
			if string(b) != `"`+key+`"` {
				panic("Key is complex. Use simpler key for: " + fmt.Sprintf("%q", key))
			}
		}
	}

	// use the production JSON formatter to format the log first. This
	// ensures JSON will marshal/unmarshal correctly in production.
	entry := hd.jsonFormatter.LogEntry(level, msg, args)

	// reset the column tracker used for fancy formatting
	col := 0

	// timestamp
	col = hd.writeColoredString(buf, entry[KeyMap.Time].(string), theme.Misc, col)

	// emphasize warnings, errors and add callstack
	context, color := hd.levelSourceContext(level, entry, args, startFrame)
	message := entry[KeyMap.Message].(string)

	// DBG, INF ...
	col = hd.set(buf, "", entry[KeyMap.Level].(string), color, col)

	// logger name
	col = hd.set(buf, "", entry[KeyMap.Name], theme.Misc, col)
	// message from user
	col = hd.set(buf, "", message, theme.Message, col)

	// Preserve key order in the sequence they were declared. This
	// makes it easier for developers to follow the log.
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
		// skip reserved keys which were already added to buffer above
		isReserved, err := isReservedKey(key)
		if err != nil {
			panic("key is invalid. Should never get here. " + err.Error())
		} else if isReserved {
			continue
		}
		col = hd.set(buf, key, entry[key], theme.Value, col)
	}

	addLF := true
	// TRC, WRN,ERR file, line number callstack
	if context != "" {
		buf.WriteRune('\n')
		addLF = context[len(context)-1:len(context)] != "\n"
		writeColor(buf, color)
		buf.WriteString(context)
		writeColor(buf, ansi.Reset)
	}

	if addLF {
		buf.WriteRune('\n')
	}

	return copyBytes(buf), nil
}

func writeColor(buf bufferWriter, code string) {
	if code == "" || disableColors {
		return
	}
	buf.WriteString(code)
}

// copyBytes makes a copy of the internal buffer slice
//
// This needs to be optimized. Misunderstood how buf.Bytes() worked.
func copyBytes(buf *bytes.Buffer) []byte {
	b := make([]byte, buf.Len())
	copy(b, buf.Bytes())
	return b
}
