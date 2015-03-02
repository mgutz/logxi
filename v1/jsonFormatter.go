package log

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// JSONFormatter formats log entries as JSON. This should be used
// in production because it is machine parseable.
type JSONFormatter struct {
	name string
}

// NewJSONFormatter creates a new instance of JSONFormatter.
func NewJSONFormatter(name string) *JSONFormatter {
	return &JSONFormatter{name: name}
}

func (jf *JSONFormatter) appendValue(buf *bytes.Buffer, val interface{}) {
	if val == nil {
		buf.WriteString("null")
		return
	}

	value := reflect.ValueOf(val)
	kind := value.Kind()
	if kind == reflect.Ptr {
		value = value.Elem()
		kind = value.Kind()
	}
	switch kind {
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

	case reflect.String:
		buf.WriteString(strconv.Quote(value.String()))

	default:
		b, err := json.Marshal(value.Interface())
		if err != nil {
			buf.WriteString("Could not json encode value:")
			buf.WriteString(err.Error())
		}
		buf.Write(b)
	}
}

func (jf *JSONFormatter) set(buf *bytes.Buffer, key string, val interface{}) {
	// WARNING: assumes this is not first key
	buf.WriteString(`, "`)
	buf.WriteString(key)
	buf.WriteString(`":`)
	jf.appendValue(buf, val)
}

// Format formats log entry as JSON.
func (jf *JSONFormatter) Format(buf *bytes.Buffer, level int, msg string, args []interface{}) {
	buf.WriteString(`{"t":"`)
	buf.WriteString(time.Now().Format("2006-01-02T15:04:05-0700"))
	buf.WriteRune('"')

	buf.WriteString(`, "l":"`)
	buf.WriteString(LevelMap[level])
	buf.WriteRune('"')

	buf.WriteString(`, "n":"`)
	buf.WriteString(jf.name)
	buf.WriteRune('"')

	buf.WriteString(`, "m":`)
	jf.appendValue(buf, msg)

	var lenArgs = len(args)
	if lenArgs > 0 {
		if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "" {
						// key is not a string, let the user know and adjust
						// for the first argument being the message
						jf.set(buf, "BADKEY"+strconv.Itoa(i+1), args[i+1])
					} else {
						jf.set(buf, key, args[i+1])
					}
				} else {
					// key is not a string, let the user know and adjust
					// for the first argument being the message
					jf.set(buf, "BADKEY"+strconv.Itoa(i+1), args[i+1])
				}
			}
		} else {
			jf.set(buf, "IMBALANCED_PAIRS", args)
		}
	}
	buf.WriteString("}\n")
}

// SetName sets the name of this formatter.
func (jf *JSONFormatter) SetName(name string) {
	jf.name = name
}
