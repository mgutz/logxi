package log

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myFloatStringer float64

func (f myFloatStringer) String() string {
	switch f {
	case myFloatStringer(math.Inf(1)):
		return "Inf"
	default:
		return fmt.Sprint(float64(f))
	}
}

func TestLogEntry_CustomScalarStringer(t *testing.T) {
	assert := assert.New(t)

	jf := NewJSONFormatter("test")

	entry := jf.LogEntry(LevelInfo, "testing myFloat", []interface{}{
		"f", myFloatStringer(10),
	})
	assert.Equal("10", entry["f"])

	entry = jf.LogEntry(LevelInfo, "testing myFloat", []interface{}{
		"f", myFloatStringer(math.Inf(1)),
	})
	assert.Equal("Inf", entry["f"])
}

type myFloatMarshaler float64

func (f myFloatMarshaler) MarshalJSON() ([]byte, error) {
	switch f {
	case myFloatMarshaler(math.Inf(1)):
		return []byte(`"Inf"`), nil
	default:
		return []byte(fmt.Sprint(float64(f))), nil
	}
}

func TestLogEntry_CustomScalarMarshaler(t *testing.T) {
	assert := assert.New(t)

	jf := NewJSONFormatter("test")

	entry := jf.LogEntry(LevelInfo, "testing myFloat", []interface{}{
		"f", myFloatMarshaler(10),
	})
	assert.Equal(float64(10), entry["f"])

	entry = jf.LogEntry(LevelInfo, "testing myFloat", []interface{}{
		"f", myFloatMarshaler(math.Inf(1)),
	})
	assert.Equal("Inf", entry["f"])
}
