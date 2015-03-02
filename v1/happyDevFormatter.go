package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/go-errors/errors"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"gopkg.in/stack.v1"
)

var markerColor func(string) string
var keyColor func(string) string
var errorColor func(string) string
var infoCode string
var warnCode string
var errorCode string
var keyCode string
var resetCode string
var plain = ""

func init() {

	ansi.DisableColors(disableColors)
	markerColor = ansi.ColorFunc("magenta")
	keyColor = ansi.ColorFunc("cyan")
	errorColor = ansi.ColorFunc("red")
	infoCode = ansi.ColorCode("green")
	warnCode = ansi.ColorCode("yellow")
	errorCode = ansi.ColorCode("red")
	keyCode = ansi.ColorCode("cyan")
	resetCode = ansi.ColorCode("reset")
	DisableColors(disableColors)
}

// DisableColors disables coloring of log entries.
func DisableColors(val bool) {
	disableColors = val
}

// GetColorableStdout gets a writer that can output colors
// on Windows and non-Widows OS. If colors are disabled,
// os.Stdout is returned.
func GetColorableStdout() io.Writer {
	if isTTY && !disableColors {
		return colorable.NewColorableStdout()
	}
	return os.Stdout
}

// HappyDevFormatter is the default recorder used if one is unspecified when
// creating a new Logger.
type HappyDevFormatter struct {
	name         string
	itoaLevelMap map[int]string
}

// NewHappyDevFormatter returns a new instance of HappyDevFormatter.
// Performance isn't priority. It's more important developers see errors
// and stack.
func NewHappyDevFormatter(name string) *HappyDevFormatter {
	var buildKV = func(level string) string {
		return Separator + keyColor("n=") + name + Separator + keyColor("l=") + level + Separator + keyColor("m=")
	}
	itoaLevelMap := map[int]string{
		LevelDebug: buildKV(LevelMap[LevelDebug]),
		LevelWarn:  buildKV(LevelMap[LevelWarn]),
		LevelInfo:  buildKV(LevelMap[LevelInfo]),
		LevelError: buildKV(LevelMap[LevelError]),
		LevelFatal: buildKV(LevelMap[LevelFatal]),
	}
	return &HappyDevFormatter{itoaLevelMap: itoaLevelMap, name: name}
}

func (tf *HappyDevFormatter) writeKey(buf *bytes.Buffer, key string) {
	// assumes this is not the first key
	buf.WriteString(Separator)
	buf.WriteString(keyCode)
	buf.WriteString(key)
	buf.WriteRune('=')
	buf.WriteString(resetCode)
}

func (tf *HappyDevFormatter) writeError(buf *bytes.Buffer, err *errors.Error) {
	buf.WriteString(errorCode)
	buf.WriteString(err.Error())
	buf.WriteRune('\n')
	buf.Write(err.Stack())
	buf.WriteString(resetCode)
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
		buf.WriteString(resetCode)
	}
}

// Format records a log entry.
func (tf *HappyDevFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	buf.WriteString(keyColor("t="))
	buf.WriteString(time.Now().Format("2006-01-02T15:04:05.000000"))

	tf.set(buf, "n", tf.name, plain)

	var colorCode string
	var context string

	switch level {
	case LevelDebug:
		colorCode = plain
	case LevelInfo:
		colorCode = infoCode
	case LevelWarn:
		c := stack.Caller(2)
		context = fmt.Sprintf("%+v", c)
		colorCode = warnCode
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
		colorCode = errorCode
	}
	tf.set(buf, "l", LevelMap[level], colorCode)
	tf.set(buf, "m", msg, colorCode)
	if context != "" {
		tf.set(buf, "c", context, colorCode)
	}

	var lenArgs = len(args)
	if lenArgs > 0 {
		fmt.Printf("lenArgs %#v\n", args)

		if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					tf.set(buf, key, args[i+1], plain)
				} else {
					tf.set(buf, "BADKEY_NAME_"+strconv.Itoa(i+1), args[i], errorCode)
					tf.set(buf, "BADKEY_VALUE_"+strconv.Itoa(i+1), args[i+1], errorCode)
				}
			}
		} else {
			buf.WriteString(errorCode)
			buf.WriteString(Separator)
			buf.WriteString("IMBALANCED_PAIRS=>")
			buf.WriteString(warnCode)
			fmt.Fprint(buf, args...)
			buf.WriteString(resetCode)
		}
	}
	buf.WriteRune('\n')
}
