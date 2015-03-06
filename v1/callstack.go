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
	if ci.lineno == 0 {
		return
	}
	start := maxInt(1, ci.lineno-ci.contextLines)
	end := ci.lineno + ci.contextLines

	f, err := os.Open(ci.filename)
	if err != nil {
		InternalLog.Error("Could not read callstack file", "file", ci.filename, "err", err)
		return
	}
	defer f.Close()

	lineno := 1
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if start <= lineno && lineno <= end {
			line := scanner.Text()
			line = expandTabs(line, 8)
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
	buf.WriteString("\t")
	buf.WriteString(ci.method)
	buf.WriteString("():")
	buf.WriteString(tildeFilename)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(ci.lineno))
	buf.WriteString("\n")

	if ci.contextLines == -1 {
		return buf.String()
	}

	for _, li := range ci.context {
		if li.lineno == ci.lineno {
			buf.WriteString(color)
			buf.WriteString(fmt.Sprintf("\t=>%5d: %s\n", li.lineno, li.line))
			continue
		}
		buf.WriteString(sourceColor)
		buf.WriteString(fmt.Sprintf("\t%7d: %s\n", li.lineno, li.line))
	}
	// get rid of last \n
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(ansi.Reset)
	return buf.String()
}
