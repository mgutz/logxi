package logxi

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullLoggerErrorOnWarn(t *testing.T) {
	testResetEnv()
	os.Setenv("LOGXI", "*=OFF")
	processEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "wrnerr", NewHappyDevFormatter("wrnerr"))

	ErrorDummy := errors.New("dummy error")

	// Warn returns error if any arg is an error type
	err := l.Warn("warn with error", "err", ErrorDummy)
	assert.Error(t, err)
	assert.Equal(t, "warn with error: dummy error", err.Error())

	// Warn returns nil error otherwise
	err = l.Warn("warn with no error", "one", 1)
	assert.NoError(t, err)
}

func TestNullLoggerErrorResult(t *testing.T) {
	testResetEnv()
	os.Setenv("LOGXI", "*=OFF")
	processEnv()
	var buf bytes.Buffer
	l := NewLogger3(&buf, "wrnerr", NewHappyDevFormatter("wrnerr"))

	ErrorDummy := errors.New("dummy error")

	// error wraps the error with pkg/errors if err type is provide
	err := l.Error("warn with error", "err", ErrorDummy)
	assert.Error(t, err)
	assert.Equal(t, "warn with error: dummy error", err.Error())

	// error returns new error based on msg otherwise
	err = l.Error("warn with no error", "one", 1)
	assert.Error(t, err)
	assert.Equal(t, "warn with no error", err.Error())
}
