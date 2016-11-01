package log

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mgutz/ansi"
)

type sourceLine struct {
	lineno int
	line   string
}

type frameInfo struct {
	filename     string
	lineno       int
	method       string
	pc           uintptr
	context      []*sourceLine
	contextLines int
}

func (ci *frameInfo) readSource(contextLines int) error {
	if ci.lineno == 0 || disableCallstack {
		return nil
	}
	start := maxInt(1, ci.lineno-contextLines)
	end := ci.lineno + contextLines

	f, err := os.Open(ci.filename)
	if err != nil {
		// if we can't read a file, it means user is running this in production
		disableCallstack = true
		return err
	}
	defer f.Close()

	lineno := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if start <= lineno && lineno <= end {
			line := scanner.Text()
			line = expandTabs(line, 4)
			ci.context = append(ci.context, &sourceLine{lineno: lineno, line: line})
		}
		lineno++
	}

	if err := scanner.Err(); err != nil {
		InternalLog.Warn("scanner error", "file", ci.filename, "err", err)
	}
	return nil
}

func (ci *frameInfo) String(color string, sourceColor string) string {
	buf := pool.Get()
	defer pool.Put(buf)

	if disableCallstack {
		buf.WriteString(color)
		buf.WriteString(Separator)
		buf.WriteString(indent)
		buf.WriteString(ci.filename)
		buf.WriteRune(':')
		buf.WriteString(strconv.Itoa(ci.lineno))
		return buf.String()
	}

	// skip anything in the logxi package
	if isLogxiCode(ci.filename) {
		return ""
	}

	// make path relative to current working directory or home
	tildeFilename, err := filepath.Rel(wd, ci.filename)
	if err != nil {
		InternalLog.Warn("Could not make path relative", "path", ci.filename)
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
	buf.WriteString(ci.method)
	buf.WriteString("(")
	buf.WriteString(tildeFilename)
	buf.WriteRune(':')
	buf.WriteString(strconv.Itoa(ci.lineno))
	buf.WriteString(")")

	if contextLines == -1 {
		return buf.String()
	}
	buf.WriteString("\n")

	// the width of the printed line number
	var linenoWidth int
	// trim spaces at start of source code based on common spaces
	var skipSpaces = 1000

	// calculate width of lineno and number of leading spaces that can be
	// removed
	for _, li := range ci.context {
		linenoWidth = maxInt(linenoWidth, len(fmt.Sprintf("%d", li.lineno)))
		index := indexOfNonSpace(li.line)
		if index > -1 && index < skipSpaces {
			skipSpaces = index
		}
	}

	for _, li := range ci.context {
		if !disableColors {
			// need to reset here.  If source is set to default color, then the message
			// color will bleed over into the source context lines
			buf.WriteString(ansi.Reset)
		}
		var format string
		format = fmt.Sprintf("%%s%%%dd:  %%s\n", linenoWidth)

		if li.lineno == ci.lineno {
			if !disableColors {
				buf.WriteString(color)
			}
			if contextLines > 0 {
				format = fmt.Sprintf("%%s=> %%%dd:  %%s\n", linenoWidth)
			}
		} else {
			if !disableColors {
				buf.WriteString(sourceColor)
			}
			if contextLines > 0 {
				// account for "=> "
				format = fmt.Sprintf("%%s%%%dd:  %%s\n", linenoWidth+3)
			}
		}
		// trim spaces at start
		idx := minInt(len(li.line), skipSpaces)
		buf.WriteString(fmt.Sprintf(format, Separator+indent+indent, li.lineno, li.line[idx:]))
	}
	// get rid of last \n
	buf.Truncate(buf.Len() - 1)
	if !disableColors {
		buf.WriteString(ansi.Reset)
	}
	return buf.String()
}

// Generates a stack from runtime.Callers()
func stackFrames(skip int, ignoreRuntime bool) []*frameInfo {
	frames := []*frameInfo{}
	size := 20
	pcs := make([]uintptr, size)
	// always skip the first frame, since it's runtime.Callers itself
	pcs = pcs[:runtime.Callers(1+skip, pcs)]

	for _, pc := range pcs {
		fn := runtime.FuncForPC(pc)
		name := fn.Name()
		file, line := fn.FileLine(pc - 1)
		if ignoreRuntime && strings.Contains(file, filepath.Join("src", "runtime")) {
			break
		}

		ci := &frameInfo{
			filename: file,
			lineno:   line,
			method:   name,
			pc:       pc,
		}

		frames = append(frames, ci)
	}
	return frames
}

// Returns debug stack excluding logxi frames
func trimmedStackTrace() string {
	buf := pool.Get()
	defer pool.Put(buf)
	frames := stackFrames(0, false)
	for _, frame := range frames {
		// skip anything in the logxi package
		if isLogxiCode(frame.filename) {
			continue
		}

		fmt.Fprintf(buf, "%s:%d (0x%x)\n", frame.filename, frame.lineno, frame.pc)

		err := frame.readSource(0)
		if err != nil || len(frame.context) < 1 {
			continue
		}

		fmt.Fprintf(buf, "\t%s: %s\n", frame.method, frame.context[0].line)
	}
	return buf.String()
}
