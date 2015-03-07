package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	. "gopkg.in/godo.v1"
)

type pair struct {
	description string
	command     string
}

var stdout io.Writer

//var promptFn = ansi.ColorFunc("cyan+h")
var promptColor = ansi.ColorCode("cyan+h")
var commentColor = ansi.ColorCode("yellow+h")
var titleColor = ansi.ColorCode("green+h")
var subtitleColor = ansi.ColorCode("black+h")
var reset = ansi.ColorCode("reset")
var normal = ansi.ColorCode("")
var wd string

func init() {
	wd, _ = os.Getwd()
	stdout = colorable.NewColorableStdout()
}

func clear() {
	Bash("clear")
	// leave a single line at top so the window
	// overlay doesn't have to be exact
	fmt.Fprintln(stdout, "")
}

func pseudoType(s string, color string) {
	if color != "" {
		fmt.Fprint(stdout, color)
	}
	for _, r := range s {
		fmt.Fprint(stdout, string(r))
		time.Sleep(50 * time.Millisecond)
	}
	if color != "" {
		fmt.Fprint(stdout, reset)
	}
}

func pseudoTypeln(s string, color string) {
	pseudoType(s, color)
	fmt.Fprint(stdout, "\n")
}

func pseudoPrompt(prompt, s string) {
	pseudoType(prompt, promptColor)
	//fmt.Fprint(stdout, promptFn(prompt))
	pseudoType(s, normal)
}

func intro(title, subtitle string, delay time.Duration) {
	clear()
	pseudoType("\n\n\t"+title+"\n\n", titleColor)
	pseudoType("\t"+subtitle, subtitleColor)
	time.Sleep(delay)
}

func typeCommand(description, commandStr string) {
	clear()
	pseudoTypeln("# "+description, commentColor)
	pseudoType("> ", promptColor)
	pseudoType(commandStr, normal)
	time.Sleep(200 * time.Millisecond)
	fmt.Fprintln(stdout, "\n")
}

var version = "v1"

func relv(p string) string {
	return filepath.Join(version, p)
}
func absv(p string) string {
	return filepath.Join(wd, version, p)
}

func tasks(p *Project) {
	p.Task("bench", func() {
		Run("LOGXI=* go test -bench . -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("build", func() {
		Run("go build", M{"$in": "v1/cmd/demo"})
	})

	p.Task("demo", func() {
		Run("go run main.go", M{"$in": "v1/cmd/demo"})
	})

	p.Task("gifcast", D{"build"}, func() {
		commands := []pair{
			{
				`create a simple app demo`,
				`cat main.ansi`,
			},
			{
				`running demo displays only warnings and errors with context`,
				`demo`,
			},
			{
				`show all log levels`,
				`LOGXI=* demo`,
			},
			{
				`enable/disable loggers with level`,
				`LOGXI=*=ERR,models,server=INF demo`,
			},
			{
				`create custom 256 colors colorscheme, pink==200`,
				`LOGXI_COLORS=*=black+h,ERR=200+b,key=blue+h demo`,
			},
			{
				`fit more on line, set time format, disable context`,
				`LOGXI=* LOGXI_FORMAT=fit,maxcol=80,t=04:05.000,context=-1 demo`,
			},
			{
				`logxi defaults to fast, unadorned JSON in production`,
				`demo | cat`,
			},
		}

		// setup time for ecorder, user presses enter when ready
		clear()
		Prompt("")

		intro(
			"log XI",
			"structured. faster. friendlier.\n\n\n\n\t::mgutz",
			1*time.Second,
		)

		for _, cmd := range commands {
			typeCommand(cmd.description, cmd.command)
			Bash(cmd.command, M{"$in": "v1/cmd/demo"})
			time.Sleep(3500 * time.Millisecond)
		}

		intro(
			"log XI demo",
			"built with godo and LICEcap ...",
			1*time.Millisecond,
		)

		Prompt("")
	})

	p.Task("demo-gif", func() {
		Bash(`cp ~/Desktop/demo.gif images`)
	})

	p.Task("allocs", func() {
		Bash(`go test -bench . -benchmem | grep "allocs\|Bench"`, M{"$in": "v1/bench"})
	}).Description("Runs benchmarks with allocs")

	p.Task("benchjson", func() {
		Bash("go test -bench=BenchmarkLoggerJSON -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("test", func() {
		Run("LOGXI=* go test", M{"$in": "v1"})
		//Run("LOGXI=* go test -run=TestColors", M{"$in": "v1"})
	})

	p.Task("isolate", D{"build"}, func() error {
		return Bash("LOGXI=* LOGXI_FORMAT=fit,maxcol=80,t=04:05.000,context=2 demo", M{"$in": "v1/cmd/demo"})
		//Run("LOGXI=* go test -run=TestWarningErrorContext", M{"$in": "v1"})
	})

	p.Task("install", func() error {
		packages := []string{
			"github.com/mattn/go-colorable",
			"github.com/mattn/go-isatty",
			"github.com/mgutz/ansi",
			"github.com/stretchr/testify/assert",

			// needed for benchmarks in bench/
			"github.com/Sirupsen/logrus",
			"gopkg.in/inconshreveable/log15.v2",
		}
		for _, pkg := range packages {
			err := Run("go get -u " + pkg)
			if err != nil {
				return err
			}
		}
		return nil
	}).Description("Installs dependencies")

}

func main() {
	Godo(tasks)
}
