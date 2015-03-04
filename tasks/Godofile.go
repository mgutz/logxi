package main

import . "gopkg.in/godo.v1"

type pair []string

func tasks(p *Project) {
	p.Task("bench", func() {
		Run("LOGXI=* go test -bench . -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("demo", func() {
		Bash(`
clear
echo github.com/mgutz/logxi
echo
echo demo built with godo and LICEcap
echo
read ok
		`)
		Run("go build", M{"$in": "v1/cmd/demo"})
		commands := []pair{
			pair{
				`default "happy" formatter shows warnings and errors`,
				`demo`,
			},
			pair{
				`Fast "JSON" formatter for production, see benchmarks`,
				`LOGXI_FORMAT=JSON demo`,
			},
			pair{
				`show all levels`,
				`LOGXI_FORMAT=* demo`,
			},
			pair{
				`show all "models" logs and only ERR for others`,
				`LOGXI=models,*=ERR demo`,
			},
			pair{
				`custom color scheme`,
				`LOGXI_COLORS=ERR=red,key=magenta,misc=white demo`,
			},
			pair{
				`focus on errors`,
				`LOGXI_COLORS=ERR=red demo`,
			},
			pair{
				`fit on line`,
				`LOGXI_FORMAT=fit,maxcol=80 demo`,
			},
			pair{
				`set custom time format`,
				`LOGXI_FORMAT=t=04:05.000000 demo`,
			},
		}

		template := `
clear

arg="# {{.description}}"
for (( i=0; i < ${#arg}; i+=1 )) ; do
	echo -n "${arg:$i:1}"
	sleep 0.1
done
sleep 0.1
echo

echo -n "\$ "
sleep 0.2
arg="{{.command}}"
for (( i=0; i < ${#arg}; i+=1 )) ; do
	echo -n "${arg:$i:1}"
	sleep 0.1
done
sleep 0.1
echo
echo

{{.command}}
sleep 3
`
		for _, cmd := range commands {
			Bash(template, M{
				"description": cmd[0],
				"command":     cmd[1],
				"$in":         "v1/cmd/demo",
			})
		}

		Bash("sleep 100")

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

	// p.Task("demo", func() {
	// 	Run("go run main.go", M{"$in": "v1/cmd/demo"})
	// }).Description("Runs the demo")
}

func main() {
	Godo(tasks)
}
