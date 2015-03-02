package log

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvLOGXI(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("LOGXI", "")
	processEnv()
	assert.Equal(LevelWarn, logxiNameLevelMap["*"], "Unset LOGXI defaults to *:WRN with TTY")

	// default all to ERR
	os.Setenv("LOGXI", "*=ERR")
	processEnv()
	level, err := getLogLevel("mylog")
	assert.NoError(err)
	assert.Equal(LevelError, level)
	level, err = getLogLevel("mylog2")
	assert.NoError(err)
	assert.Equal(LevelError, level)

	// unrecognized defaults to LevelDebug on TTY
	os.Setenv("LOGXI", "mylog=badlevel")
	processEnv()
	level, err = getLogLevel("mylog")
	assert.Equal(LevelWarn, level)

	// wildcard should not override exact match
	os.Setenv("LOGXI", "*=WRN,mylog=ERR,other=OFF")
	processEnv()
	level, err = getLogLevel("mylog")
	assert.Equal(LevelError, level)
	level, err = getLogLevel("other")
	assert.Error(err, "OFF should return error")

	// wildcard pattern should match
	os.Setenv("LOGXI", "*log=ERR")
	processEnv()
	level, err = getLogLevel("mylog")
	assert.Equal(LevelError, level, "wildcat prefix should match")

	os.Setenv("LOGXI", "myx*=ERR")
	processEnv()
	level, err = getLogLevel("mylog")
	assert.Error(err, "no match should return error")

	os.Setenv("LOGXI", "myl*,-foo")
	processEnv()
	level, err = getLogLevel("mylog")
	assert.NoError(err)
	assert.Equal(LevelDebug, level)
	level, err = getLogLevel("foo")
	assert.Error(err)
}

func TestEnvLOGXI_FORMAT(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("LOGXI_FORMAT", "")
	processEnv()
	assert.Equal(FormatText, logxiFormat, "TTY defaults to FormatText")

	os.Setenv("LOGXI_FORMAT", "JSON")
	processEnv()
	assert.Equal(FormatJSON, logxiFormat)

	os.Setenv("LOGXI_FORMAT", "json")
	processEnv()
	assert.Equal(FormatText, logxiFormat, "Mismatches defaults to FormatText")
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

func testResetEnv() {
	os.Setenv("LOGXI", "")
	//os.Setenv("LOGXI_COLORS", "")
	os.Setenv("LOGXI_FORMAT", "")
	processEnv()
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
	assert.Exactly(t, []interface{}{"foo"}, obj["IMBALANCED_PAIRS"])
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
