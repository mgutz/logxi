package log

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
	relFilename  string
	lineno       int
	method       string
	context      []*sourceLine
	contextLines int
}

func newCallstackInfo(callstack interface{}, contextLines int) *callstackInfo {
	filename := fmt.Sprintf("%#s", callstack)
	relFilename := fmt.Sprintf("%+s", callstack)
	linestr := fmt.Sprintf("%d", callstack)
	lineno, _ := strconv.Atoi(linestr)
	fnname := fmt.Sprintf("%n", callstack)
	ci := &callstackInfo{
		filename:     filename,
		relFilename:  relFilename,
		lineno:       lineno,
		method:       fnname,
		context:      []*sourceLine{},
		contextLines: contextLines,
	}
	ci.readSource()
	return ci
}

func (ci *callstackInfo) readSource() {
	if ci.lineno == 0 || disableCallstack {
		return
	}
	start := maxInt(1, ci.lineno-ci.contextLines)
	end := ci.lineno + ci.contextLines

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

func (ci *callstackInfo) dump() {
	fmt.Printf("DBG: HERE\n")
	fmt.Printf("ci.filename %#v\n", ci.filename)
	fmt.Printf("ci.lineno %#v\n", ci.lineno)
	fmt.Printf("ci.method %#v\n", ci.method)
	fmt.Printf("ci.relFilename %#v\n", ci.relFilename)
	fmt.Printf("first", !rePackageTestFile.MatchString(ci.filename))
	fmt.Printf("second", rePackageFile.MatchString(ci.relFilename))
}

func (ci *callstackInfo) String(color string, sourceColor string) string {
	// skip anything in the logxi package (except for tests)
	if !rePackageTestFile.MatchString(ci.filename) && rePackageFile.MatchString(ci.relFilename) {
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
	buf.WriteString("at ")
	buf.WriteString(ci.method)
	buf.WriteString("(")
	buf.WriteString(tildeFilename)
	buf.WriteString("):")
	buf.WriteString(strconv.Itoa(ci.lineno))

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
