package logxi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

type bufferWriter interface {
	Write(p []byte) (nn int, err error)
	WriteByte(byte) error
	WriteRune(r rune) (n int, err error)
	WriteString(s string) (n int, err error)
}

// JSONFormatter is a fast, efficient JSON formatter optimized for logging.
//
// * log entry keys are not escaped
//   Who uses complex keys when coding? Checked by HappyDevFormatter in case user does.
//   Nested object keys are escaped by json.Marshal().
// * Primitive types uses strconv
// * Logger reserved key values (time, log name, level) require no conversion
// * sync.Pool buffer for bytes.Buffer
type JSONFormatter struct {
	name string
}

// NewJSONFormatter creates a new instance of JSONFormatter.
func NewJSONFormatter(name string) *JSONFormatter {
	return &JSONFormatter{name: name}
}

func (jf *JSONFormatter) writeString(buf bufferWriter, s string) {
	b, err := json.Marshal(s)
	if err != nil {
		InternalLog.Error("Could not json.Marshal string.", "str", s)
		buf.WriteString(`"Could not marshal this key's string"`)
		return
	}
	buf.Write(b)
}

func (jf *JSONFormatter) writeError(buf bufferWriter, err error) {
	jf.writeString(buf, err.Error())
	stack := callstack(7, -1, false)
	b, err := json.Marshal(stack)
	if err != nil {
		InternalLog.Error("Could not json.Marshal callstack.", "callstack", stack)
	} else {
		jf.set(buf, KeyMap.CallStack, b)
	}
}

func (jf *JSONFormatter) appendValue(buf bufferWriter, val interface{}) {
	if val == nil {
		buf.WriteString("null")
		return
	}

	switch T := val.(type) {
	case []byte:
		buf.Write(T)
		return
	case error:
		// always show error stack even at cost of some performance. there's
		// nothing worse than looking at production logs without a clue
		jf.writeError(buf, T)
		return
	}

	value := reflect.ValueOf(val)
	kind := value.Kind()
	if kind == reflect.Ptr {
		if value.IsNil() {
			buf.WriteString("null")
			return
		}
		value = value.Elem()
		kind = value.Kind()
	}
	switch kind {
	case reflect.String:
		appendString(buf, value.String())

	case reflect.Bool:
		if value.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(value.Int(), 10))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(value.Uint(), 10))

	case reflect.Float32:
		buf.WriteString(strconv.FormatFloat(value.Float(), 'g', -1, 32))

	case reflect.Float64:
		buf.WriteString(strconv.FormatFloat(value.Float(), 'g', -1, 64))

	default:
		var err error
		var b []byte
		if stringer, ok := val.(fmt.Stringer); ok {
			b, err = json.Marshal(stringer.String())
		} else {
			b, err = json.Marshal(val)
		}

		if err != nil {
			InternalLog.Error("Could not json.Marshal value: ", "formatter", "JSONFormatter", "err", err.Error())
			b, err = json.Marshal(fmt.Sprintf("%#v", val))
			if err != nil {
				// should never get here, but JSONFormatter should never panic
				msg := "Could not Sprintf value"
				InternalLog.Error(msg)
				buf.WriteString(`"` + msg + `"`)
				return
			}
		}
		buf.Write(b)
	}
}

func (jf *JSONFormatter) set(buf bufferWriter, key string, val interface{}) {
	// WARNING: assumes this is not first key
	buf.WriteString(`, "`)
	buf.WriteString(key)
	buf.WriteString(`":`)
	jf.appendValue(buf, val)
}

// Format formats log entry as JSON.
func (jf *JSONFormatter) Format(level int, msg string, args []interface{}, startFrame int) ([]byte, error) {
	buf := pool.Get()
	defer pool.Put(buf)

	const lead = `", "`
	const colon = `":"`

	buf.WriteString(`{"`)
	buf.WriteString(KeyMap.Time)
	buf.WriteString(`":"`)
	buf.WriteString(time.Now().Format(timeFormat))

	buf.WriteString(`", "`)
	buf.WriteString(KeyMap.PID)
	buf.WriteString(`":"`)
	buf.WriteString(pidStr)

	buf.WriteString(`", "`)
	buf.WriteString(KeyMap.Level)
	buf.WriteString(`":"`)
	buf.WriteString(LevelMap[level])

	buf.WriteString(`", "`)
	buf.WriteString(KeyMap.Name)
	buf.WriteString(`":"`)
	buf.WriteString(jf.name)

	buf.WriteString(`", "`)
	buf.WriteString(KeyMap.Message)
	buf.WriteString(`":`)
	jf.appendValue(buf, msg)

	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs == 1 {
			jf.set(buf, singleArgKey, args[0])
		} else if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "" {
						// show key is invalid
						jf.set(buf, badKeyAtIndex(i), args[i+1])
					} else {
						jf.set(buf, key, args[i+1])
					}
				} else {
					// show key is invalid
					jf.set(buf, badKeyAtIndex(i), args[i+1])
				}
			}
		} else {
			jf.set(buf, warnImbalancedKey, args)
		}
	}

	buf.WriteString("}\n")
	return copyBytes(buf), nil
}

// LogEntry returns the JSON log entry object built by Format(). Used by
// HappyDevFormatter to ensure any data logged while developing properly
// logs in production.
func (jf *JSONFormatter) LogEntry(level int, msg string, args []interface{}) map[string]interface{} {
	b, err := jf.Format(level, msg, args, 0)
	if err != nil {
		panic("Unable to format entry from JSONFormatter: " + err.Error() + " \"" + string(b) + "\"")
	}
	var entry map[string]interface{}
	err = json.Unmarshal(b, &entry)
	if err != nil {
		panic("Unable to unmarshal entry from JSONFormatter: " + err.Error() + " \"" + string(b) + "\"")
	}
	return entry
}

const _hex = "0123456789abcdef"

// CREDIT TO https://raw.githubusercontent.com/uber-go/zap/master/json_encoder.go

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's escaping function, it doesn't attempt to
// protect the user from browser vulnerabilities or JSONP-related problems.
func appendString(buf bufferWriter, s string) {
	buf.WriteRune('"')
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '\\' && b != '"' {
				buf.WriteByte(b)
				continue
			}
			switch b {
			case '\\', '"':
				buf.WriteRune('\\')
				buf.WriteByte(b)
			case '\n':
				buf.WriteRune('\\')
				buf.WriteRune('n')
			case '\r':
				buf.WriteRune('\\')
				buf.WriteRune('r')
			case '\t':
				buf.WriteRune('\\')
				buf.WriteRune('t')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				buf.WriteString(`\u00`)
				buf.WriteByte(_hex[b>>4])
				buf.WriteByte(_hex[b&0xF])
			}
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			buf.WriteString(`\ufffd`)
			i++
			continue
		}
		buf.WriteString(s[i : i+size])
		i += size
	}
	buf.WriteRune('"')
}
