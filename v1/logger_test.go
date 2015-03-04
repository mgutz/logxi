package log

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func processEnv() {
	ProcessEnv(readFromEnviron())
}

func testResetEnv() {
	os.Clearenv()
	processEnv()
}

func TestEnvLOGXI(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("LOGXI", "")
	processEnv()
	assert.Equal(LevelWarn, logxiNameLevelMap["*"], "Unset LOGXI defaults to *:WRN with TTY")

	// default all to ERR
	os.Setenv("LOGXI", "*=ERR")
	processEnv()
	level := getLogLevel("mylog")
	assert.Equal(LevelError, level)
	level = getLogLevel("mylog2")
	assert.Equal(LevelError, level)

	// unrecognized defaults to LevelDebug on TTY
	os.Setenv("LOGXI", "mylog=badlevel")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelWarn, level)

	// wildcard should not override exact match
	os.Setenv("LOGXI", "*=WRN,mylog=ERR,other=OFF")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelError, level)
	level = getLogLevel("other")
	assert.Equal(LevelOff, level)

	// wildcard pattern should match
	os.Setenv("LOGXI", "*log=ERR")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelError, level, "wildcat prefix should match")

	os.Setenv("LOGXI", "myx*=ERR")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelOff, level, "no match should return off")

	os.Setenv("LOGXI", "myl*,-foo")
	processEnv()
	level = getLogLevel("mylog")
	assert.Equal(LevelDebug, level)
	level = getLogLevel("foo")
	assert.Equal(LevelOff, level)
}

func TestEnvLOGXI_FORMAT(t *testing.T) {
	assert := assert.New(t)
	oldIsTerminal := isTerminal

	os.Setenv("LOGXI_FORMAT", "")
	setDefaults(true)
	processEnv()
	assert.Equal(FormatHappy, logxiFormat, "terminal defaults to FormatHappy")
	setDefaults(false)
	processEnv()
	assert.Equal(FormatText, logxiFormat, "non terminal defaults to FormatText")

	os.Setenv("LOGXI_FORMAT", "JSON")
	processEnv()
	assert.Equal(FormatJSON, logxiFormat)

	os.Setenv("LOGXI_FORMAT", "json")
	setDefaults(true)
	processEnv()
	assert.Equal(FormatHappy, logxiFormat, "Mismatches defaults to FormatHappy")
	setDefaults(false)
	processEnv()
	assert.Equal(FormatText, logxiFormat, "Mismatches defaults to FormatText non terminal")

	isTerminal = oldIsTerminal
	setDefaults(isTerminal)
}

func TestColors(t *testing.T) {
	testResetEnv()
	l := New("bench")
	l.SetLevel(LevelDebug)
	l.Debug("just another day", "key")
	l.Debug("and another one", "key")
	l.Info("something you should know")
	l.Warn("hmm didn't expect that")
	l.Error("oh oh, you're in trouble", "key", 1)
}

func TestJSON(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "bench")
	l.SetLevel(LevelDebug)
	l.SetFormatter(NewJSONFormatter("bench"))
	l.Error("hello", "foo", "bar")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj["foo"].(string))
	assert.Equal(t, "hello", obj["m"].(string))
}

func TestJSONImbalanced(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "bench")
	l.SetLevel(LevelDebug)
	l.SetFormatter(NewJSONFormatter("bench"))
	l.Error("hello", "foo")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Exactly(t, []interface{}{"foo"}, obj[warnImbalancedKey])
	assert.Equal(t, "hello", obj["m"].(string))
}

func TestJSONNoArgs(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "bench")
	l.SetLevel(LevelDebug)
	l.SetFormatter(NewJSONFormatter("bench"))
	l.Error("hello")

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "hello", obj["m"].(string))
}

func TestJSONNested(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "bench")
	l.SetLevel(LevelDebug)
	l.SetFormatter(NewJSONFormatter("bench"))
	l.Error("hello", "obj", map[string]string{"fruit": "apple"})

	var obj map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "hello", obj["m"].(string))
	o := obj["obj"]
	assert.Equal(t, "apple", o.(map[string]interface{})["fruit"].(string))
}

func TestParseLogEnvError(t *testing.T) {
	testResetEnv()
	os.Setenv("LOGXI", "ERR=red")
	processLogEnv(readFromEnviron())
}

func TestKeyNotString(t *testing.T) {
	testResetEnv()
	var buf bytes.Buffer
	l := NewLogger(&buf, "badkey")
	l.SetLevel(LevelDebug)
	l.SetFormatter(NewHappyDevFormatter("badkey"))
	l.Debug("foo", 1)
	assert.Panics(t, func() {
		l.Debug("reserved key", "t", "trying to use time")
	})
}
