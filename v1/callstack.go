package log

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mgutz/ansi"
)

type sourceLine struct {
	lineno int
	line   string
}

type callstackInfo struct {
	filename     string
	lineno       int
	method       string
	context      []*sourceLine
	contextLines int
}

func newCallstackInfo(callstack interface{}, contextLines int) *callstackInfo {
	filename := fmt.Sprintf("%#s", callstack)
	linestr := fmt.Sprintf("%d", callstack)
	lineno, _ := strconv.Atoi(linestr)
	fnname := fmt.Sprintf("%n", callstack)
	ci := &callstackInfo{
		filename: filename,
		lineno:   lineno,
		method:   fnname,
		context:  []*sourceLine{},
	}
	ci.readSource(contextLines)
	return ci
}

func (ci *callstackInfo) readSource(contextLines int) {
	if ci.lineno == 0 || disableCallstack {
		return
	}
	start := maxInt(1, ci.lineno-contextLines)
	end := ci.lineno + contextLines

	f, err := os.Open(ci.filename)
	if err != nil {
		// if we can't read a file, it means user is running this in production
		disableCallstack = true
		InternalLog.Error("Disabling callstack context. Maybe you are running in production? Could not read source file.", "file", ci.filename, "err", err)
		return
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
		InternalLog.Error("scanner error", "file", ci.filename, "err", err)
	}
}

var rePackageFile = regexp.MustCompile(`logxi/v1/\w+\.go`)
var rePackageTestFile = regexp.MustCompile(`logxi/v1/\w+_test\.go`)

func (ci *callstackInfo) String(color string, sourceColor string) string {
	// skip anything in the logxi package (except for tests)
	if !rePackageTestFile.MatchString(ci.filename) && rePackageFile.MatchString(ci.filename) {
		return ""
	}

	tildeFilename := ci.filename
	if strings.HasPrefix(ci.filename, wd) {
		tildeFilename = strings.Replace(ci.filename, wd+string(os.PathSeparator), "", 1)
	} else if strings.HasPrefix(ci.filename, home) {
		tildeFilename = strings.Replace(ci.filename, home, "~", 1)
	}

	var buf bytes.Buffer
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

	if ci.contextLines == -1 {
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
		var format string
		format = fmt.Sprintf("%%s%%%dd:  %%s\n", linenoWidth)

		if li.lineno == ci.lineno {
			buf.WriteString(color)
			if ci.contextLines > 2 {
				format = fmt.Sprintf("%%s=> %%%dd:  %%s\n", linenoWidth)
			}
		} else {
			buf.WriteString(sourceColor)
			if ci.contextLines > 2 {
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
	buf.WriteString(ansi.Reset)
	return buf.String()
}

func parseDebugStack(stack string, skip int, ignoreRuntime bool) []*callstackInfo {
	lines := strings.Split(stack, "\n")
	frames := []*callstackInfo{}
	for i := skip * 2; i < len(lines); i += 2 {
		ci := &callstackInfo{}
		sourceLine := lines[i]
		if sourceLine == "" {
			break
		}
		if ignoreRuntime && strings.Contains(sourceLine, filepath.Join("src", "runtime")) {
			break
		}

		colon := strings.Index(sourceLine, ":")
		space := strings.Index(sourceLine, " ")
		ci.filename = sourceLine[0:colon]
		numstr := sourceLine[colon+1 : space]
		lineno, err := strconv.Atoi(numstr)
		if err != nil {
			InternalLog.Error("Could not parse line number", "sourceLine", sourceLine, "numstr", numstr)
			continue
		}
		ci.lineno = lineno

		methodLine := lines[i+1]
		colon = strings.Index(methodLine, ":")
		ci.method = strings.Trim(methodLine[0:colon], "\t ")
		frames = append(frames, ci)
	}
	return frames
}

func parseLogxiStack(entry map[string]interface{}, skip int, ignoreRuntime bool) []*callstackInfo {
	var errStack string
	// logxi.stack:connection error
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/jsonFormatter.go:45 (0x5fa70)
	// 	(*JSONFormatter).writeError: jf.writeString(buf, err.Error()+"\n"+string(debug.Stack()))
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/jsonFormatter.go:82 (0x5fdc3)
	// 	(*JSONFormatter).appendValue: jf.writeError(buf, err)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/jsonFormatter.go:109 (0x605ca)
	// 	(*JSONFormatter).set: jf.appendValue(buf, val)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/jsonFormatter.go:138 (0x60a2c)
	// 	(*JSONFormatter).Format: jf.set(buf, key, args[i+1])
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/jsonFormatter.go:157 (0x60c10)
	// 	(*JSONFormatter).LogEntry: jf.Format(&buf, level, msg, args)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/happyDevFormatter.go:263 (0x5ddb8)
	// 	(*HappyDevFormatter).Format: entry := hd.jsonFormatter.LogEntry(level, msg, args)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/defaultLogger.go:87 (0x5a4a9)
	// 	(*DefaultLogger).Log: l.formatter.Format(&buf, level, msg, args)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/defaultLogger.go:75 (0x5a343)
	// 	(*DefaultLogger).Error: l.Log(LevelError, msg, args)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/cmd/demo/main.go:15 (0x214c)
	// 	causeError: logger.Error("error in function", "err", errConnection)
	// /Users/mgutz/go/src/github.com/mgutz/logxi/v1/cmd/demo/main.go:31 (0x2503)
	// 	main: causeError()
	// /Users/mgutz/goroot/src/runtime/proc.go:63 (0x13d13)
	// 	main: main_main()
	// /Users/mgutz/goroot/src/runtime/asm_amd64.s:2232 (0x38bf1)
	// 	goexit:

	found := ""
	for key, val := range entry {
		if s, ok := val.(string); ok {
			if strings.HasPrefix(s, "logxi.stack") {
				found = key
				errStack = s
				break
			}
		}
	}
	if found == "" {
		return nil
	}

	// get rid of "logxi.stack:"
	colon := strings.IndexRune(errStack, ':')
	newline := strings.IndexRune(errStack, '\n')
	msg := errStack[colon+1 : newline]
	frames := parseDebugStack(errStack[newline+1:], skip, ignoreRuntime)

	// replace original message with just the error
	entry[found] = msg
	return frames
}
