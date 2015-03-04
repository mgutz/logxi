package main

import (
	"fmt"
	"io"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	. "gopkg.in/godo.v1"
)

type pair []string

var stdout io.Writer
var promptFn = ansi.ColorFunc("cyan+h")
var commentColor = ansi.ColorCode("yellow+h")
var subduedColor = ansi.ColorCode("black+h")
var reset = ansi.ColorCode("reset")

func init() {
	stdout = colorable.NewColorableStdout()
}

func clear() {
	Bash("clear")
	// leave a single line at top so the window
	// overlay doesn't have to be exact
	fmt.Fprintln(stdout, "")
}

func pseudoType(s string) {
	for _, r := range s {
		fmt.Fprint(stdout, string(r))
		time.Sleep(50 * time.Millisecond)
	}
}

func pseudoSubdued(s string) {
	fmt.Fprint(stdout, subduedColor)
	pseudoType(s)
	fmt.Fprint(stdout, reset)
}

func pseudoComment(s string) {
	fmt.Fprint(stdout, commentColor)
	pseudoType(s)
	fmt.Fprint(stdout, reset)
}

func pseudoPrompt(prompt, s string) {
	fmt.Fprint(stdout, promptFn(prompt))
	pseudoType(s)
}

func intro(title, subtitle string) {
	clear()
	pseudoType("\n\n\t" + title + "\n\n")
	pseudoSubdued("\t" + subtitle)
	Prompt("")
}

func tasks(p *Project) {
	p.Task("bench", func() {
		Run("LOGXI=* go test -bench . -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("demo", func() {
		clear()
		Prompt("")
		intro(
			"log XI - faster, friendlier and more awesome",
			"built with godo and licecap ...",
		)

		Run("go build", M{"$in": "v1/cmd/demo"})
		commands := []pair{
			pair{
				`default "happy" formatter shows warnings and errors, aligns keys`,
				`demo`,
			},
			pair{
				`Fast "JSON" formatter for production, see benchmarks`,
				`LOGXI_FORMAT=JSON demo`,
			},
			pair{
				`show all log levels`,
				`LOGXI=* demo`,
			},
			pair{
				`show all "models" logs and only errors and above for others`,
				`LOGXI=models,*=ERR demo`,
			},
			pair{
				`custom color scheme can be put in your bashrc/zshrc`,
				`LOGXI_COLORS=ERR=red,key=magenta,misc=white demo`,
			},
			pair{
				`too many rainbows? just errors`,
				`LOGXI_COLORS=ERR=red demo`,
			},
			pair{
				`fit more info on line`,
				`LOGXI_FORMAT=fit,maxcol=80 demo`,
			},
			pair{
				`set custom time format`,
				`LOGXI_FORMAT=t=04:05.000 demo`,
			},
		}

		for _, cmd := range commands {
			clear()
			pseudoComment("# " + cmd[0] + "\n")
			pseudoPrompt("> ", cmd[1])
			time.Sleep(200 * time.Millisecond)
			fmt.Fprintln(stdout, "\n")
			Bash(cmd[1], M{"$in": "v1/cmd/demo"})
			time.Sleep(3 * time.Second)
		}
		clear()
		fmt.Println("\n\n\tlog XI by @mgutz\n")
		time.Sleep(1 * time.Minute)
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
