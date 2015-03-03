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
		//fmt.Printf("[%s] %s=%q\n", key, style, c)
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

// GetColorableStdout gets a writer that can output colors
// on Windows and non-Widows OS. If colors are disabled,
// os.Stdout is returned.
// func GetColorableStdout() io.Writer {
// 	if isTerminal && !disableColors {
// 		fmt.Println("returning Newcolorable")
// 		return colorable.NewColorableStdout()
// 	}
// 	fmt.Println("returning stdout")
// 	return os.Stdout
// }

// HappyDevFormatter is the formatter used for terminals. It is
// colorful, dev friendly and provides meaningful logs when
// warnings and errors occur. DO NOT use in production
type HappyDevFormatter struct {
	name string
}

// NewHappyDevFormatter returns a new instance of HappyDevFormatter.
// Performance isn't priority. It's more important developers see errors
// and stack.
func NewHappyDevFormatter(name string) *HappyDevFormatter {
	return &HappyDevFormatter{name: name}
}

func (tf *HappyDevFormatter) writeKey(buf *bytes.Buffer, key string) {
	// assumes this is not the first key
	buf.WriteString(Separator)
	if key == "" {
		return
	}
	buf.WriteString(theme.Key)
	buf.WriteString(key)
	buf.WriteRune('=')
	buf.WriteString(theme.Reset)
}

func (tf *HappyDevFormatter) writeError(buf *bytes.Buffer, err *errors.Error) {
	buf.WriteString(theme.Error)
	buf.WriteString(err.Error())
	buf.WriteRune('\n')
	buf.Write(err.Stack())
	buf.WriteString(theme.Reset)
}

func (tf *HappyDevFormatter) set(buf *bytes.Buffer, key string, value interface{}, colorCode string) {
	tf.writeKey(buf, key)
	if colorCode != "" {
		buf.WriteString(colorCode)
	}
	if err, ok := value.(error); ok {
		err2 := errors.Wrap(err, 4)
		tf.writeError(buf, err2)
	} else if err, ok := value.(*errors.Error); ok {
		tf.writeError(buf, err)
	} else {
		fmt.Fprintf(buf, "%v", value)
	}
	if colorCode != "" {
		buf.WriteString(theme.Reset)
	}
}

// Format records a log entry.
func (tf *HappyDevFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	buf.WriteString(theme.Misc)
	buf.WriteString(time.Now().Format(timeFormat))
	buf.WriteString(theme.Reset)

	var colorCode string
	var context string

	switch level {
	case LevelDebug:
		colorCode = theme.Debug
	case LevelInfo:
		colorCode = theme.Info
	case LevelWarn:
		c := stack.Caller(2)
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
	// tf.set(buf, "l", LevelMap[level], colorCode)
	// tf.set(buf, "n", tf.name, theme.Value)
	// tf.set(buf, "m", msg, colorCode)
	tf.set(buf, "", LevelMap[level], colorCode)
	tf.set(buf, "", tf.name, theme.Misc)
	tf.set(buf, "", msg, colorCode)

	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					tf.set(buf, key, args[i+1], theme.Value)
				} else {
					tf.set(buf, "BADKEY_NAME_"+strconv.Itoa(i+1), args[i], theme.Error)
					tf.set(buf, "BADKEY_VALUE_"+strconv.Itoa(i+1), args[i+1], theme.Error)
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
	if context != "" {
		tf.set(buf, "@", context, colorCode)
	}
	buf.WriteRune('\n')
	buf.WriteString(theme.Reset)
}
