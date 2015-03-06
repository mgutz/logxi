package log

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBuf bytes.Buffer

var testInternalLog Logger

func init() {
	testInternalLog = NewLogger(&testBuf, "__logxi")
	testInternalLog.SetLevel(LevelError)
	testInternalLog.SetFormatter(NewTextFormatter("__logxi"))
}

func TestUnknownLevel(t *testing.T) {
	testResetEnv()
	os.Setenv("LOGXI", "*=oy")
	processEnv()
	buffer := testBuf.String()
	assert.Contains(t, buffer, "Unknown level", "should error on unknown level")
}
