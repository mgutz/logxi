package log

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// RFC5424Formatter is the default recorder used if one is unspecified when
// creating a new Logger.
type RFC5424Formatter struct {
	name string
}

// NewRFC5424Formatter returns a new instance of RFC5424Formatter. SetName
// must be called befored using it.
func NewRFC5424Formatter(name string) *RFC5424Formatter {
	return &RFC5424Formatter{name: name}
}

func (tf *RFC5424Formatter) set(buf bufferWriter, key string, val interface{}) {
	buf.WriteString(" ")
	buf.WriteString(key)
	buf.WriteString("=")
	if err, ok := val.(error); ok {
		buf.WriteString(err.Error())
		buf.WriteString(strings.Replace(string(debug.Stack()), "\n", " | ", -1))
		return
	}
	buf.WriteString(fmt.Sprintf("\"%v\"", val))
}

// Format records a log entry.
func (tf *RFC5424Formatter) Format(writer io.Writer, level int, msg string, args []interface{}) {
	buf := pool.Get()
	defer pool.Put(buf)

	// <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROC-ID MSGID MSG"

	// Prival = Facility + Priority
	// Facility is "user", so "1", x 8.  Therefore, 8
	// Priority is the same as level, except Emergency is 0 not -1, and there is
	// no Trace or All so these are converted to DEBUG
	priority := level
	if level < 0 {
		priority = 0
	} else if level > 7 {
		priority = 7
	}
	prival := 8 + priority

	buf.WriteString("<" + strconv.Itoa(prival) + ">1 ")
	buf.WriteString(time.Now().Format(time.RFC3339))
	buf.WriteString(" ")
	hostname, _ := os.Hostname()
	buf.WriteString(hostname)
	buf.WriteString(" ")
	buf.WriteString(tf.name)
	buf.WriteString(" ")
	buf.WriteString(pidStr)
	buf.WriteString(" - ")

	var lenArgs = len(args)
	if lenArgs > 0 {
		buf.WriteString("[SDID@1")
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
		buf.WriteString("] ")
	}
	buf.WriteString(msg)
	buf.WriteRune('\n')
	buf.WriteTo(writer)
}
