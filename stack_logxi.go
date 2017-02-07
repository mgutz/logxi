package logxi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mgutz/ansi"
)

// Name is the full package and function name.
func (f Frame) Name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return ""
	}
	return fn.Name()
}

// Method is the method name derived from Name.
func (f Frame) Method() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return ""
	}
	name := fn.Name()

	if idx := strings.LastIndex(name, "/"); idx > 0 {
		return name[idx+1:]
	}

	return name
}

// MarshalJSON implements JSON marshaller
func (f Frame) MarshalJSON() ([]byte, error) {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return nil, nil
	}

	file, line := fn.FileLine(f.pc())
	return json.Marshal(fmt.Sprintf("%s() %s:%d", fn.Name(), file, line))
}

func (f Frame) String(color string, sourceColor string) string {
	buf := pool.Get()
	defer pool.Put(buf)

	file, line, method := f.File(), f.Line(), f.Method()

	if disableCallstack {
		buf.WriteString(color)
		buf.WriteString(Separator)
		buf.WriteString(indent)
		buf.WriteString(file)
		buf.WriteRune(':')
		buf.WriteString(strconv.Itoa(line))
		return buf.String()
	}

	// make path relative to current working directory or home
	tildeFilename, err := filepath.Rel(wd, file)
	if err != nil {
		InternalLog.Warn("Could not make path relative", "path", file)
		return ""
	}
	// ../../../ is too complex.  Make path relative to home
	if strings.HasPrefix(tildeFilename, strings.Repeat(".."+string(os.PathSeparator), 3)) {
		tildeFilename = strings.Replace(tildeFilename, home, "~", 1)
	}

	buf.WriteString(color)
	buf.WriteString(Separator)
	buf.WriteString(indent)
	buf.WriteString("in ")
	buf.WriteString(method)
	buf.WriteString("() ")
	buf.WriteString(tildeFilename)
	buf.WriteRune(':')
	buf.WriteString(strconv.Itoa(line))

	if contextLines == -1 {
		return buf.String()
	}
	buf.WriteString("\n")

	// the width of the printed line number
	var linenoWidth int
	// trim spaces at start of source code based on common spaces
	var skipSpaces = 1000

	const showArrowThreshold = 0

	sourceContext, err := readSourceContext(file, line)
	if err != nil {
		InternalLog.Warn("Could not get source context", "path", file)
		return ""
	}

	// calculate width of lineno and number of leading spaces that can be
	// removed
	for _, li := range sourceContext {
		linenoWidth = maxInt(linenoWidth, len(fmt.Sprintf("%d", li.line)))
		index := indexOfNonSpace(li.text)
		if index > -1 && index < skipSpaces {
			skipSpaces = index
		}
	}

	for _, li := range sourceContext {
		var format string
		format = fmt.Sprintf("%%s%%%dd:  %%s\n", linenoWidth)

		if li.line == line {
			buf.WriteString(color)
			if contextLines > showArrowThreshold {
				format = fmt.Sprintf("%%s=> %%%dd:  %%s\n", linenoWidth)
			}
		} else {
			buf.WriteString(sourceColor)
			if contextLines > showArrowThreshold {
				// account for "=> "
				format = fmt.Sprintf("%%s%%%dd:  %%s\n", linenoWidth+3)
			}
		}
		// trim spaces at start
		idx := minInt(len(li.text), skipSpaces)
		buf.WriteString(fmt.Sprintf(format, Separator+indent+indent, li.line, li.text[idx:]))
	}
	// get rid of last \n
	buf.Truncate(buf.Len() - 1)
	if !disableColors {
		buf.WriteString(ansi.Reset)
	}
	return buf.String()
}

// Gets a slice of the callstack. Set size to -1 to retrieve all.
func callstack(skip int, size int, useIgnore bool) StackTrace {
	var frames []Frame
	if useIgnore {
		for _, f := range callers().StackTrace() {
			if !isIgnored(f) {
				frames = append(frames, f)
			}
		}
	} else {
		frames = callers().StackTrace()
	}

	L := len(frames)

	if skip > L-1 {
		return []Frame{}
	}

	if size < 0 || skip+size > L {
		return frames[skip:]
	}

	return frames[skip : skip+size]
}

type sourceLine struct {
	line int
	text string
}

func readSourceContext(file string, line int) ([]sourceLine, error) {
	if line == 0 || disableCallstack {
		return nil, nil
	}

	var result []sourceLine

	start := maxInt(1, line-contextLines)
	end := line + contextLines

	f, err := os.Open(file)
	if err != nil {
		InternalLog.Error("Could not open file", "file", file)
		// if we can't read a file, it means user is running this in production
		disableCallstack = true
		return nil, nil
	}
	defer f.Close()

	lineno := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if start <= lineno && lineno <= end {
			text := scanner.Text()
			text = expandTabs(text, 4)
			result = append(result, sourceLine{line: lineno, text: text})
		}
		lineno++
	}

	if err := scanner.Err(); err != nil {
		InternalLog.Warn("scanner error", "file", file, "err", err)
	}
	return result, nil
}
