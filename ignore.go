package logxi

import (
	"path/filepath"
	"strings"
)

var inLogxiPath = filepath.Join("github.com", "mgutz", "logxi")
var inRuntimePath = filepath.Join("go", "src", "runtime")
var ignoreFuncs = []func(f Frame) bool{defaultIgnore}

func isIgnored(f Frame) bool {
	for _, fn := range ignoreFuncs {
		if fn(f) {
			return true
		}
	}
	return false
}

func defaultIgnore(f Frame) bool {
	filename := f.File()
	dirname := filepath.Dir(filename)
	// need to see errors in tests
	ignored := (strings.HasSuffix(dirname, inLogxiPath) && !strings.HasSuffix(filename, "_test.go")) ||
		strings.HasSuffix(dirname, inRuntimePath)
	return ignored
}

// AddIgnoreFilter adds an ignore function which is used to exclude frames from callstack.
func AddIgnoreFilter(fn func(f Frame) bool) {
	ignoreFuncs = append(ignoreFuncs, fn)
}
