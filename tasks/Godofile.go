package main

import . "gopkg.in/godo.v1"

func tasks(p *Project) {
	p.Task("bench", func() {
		Run("LOGXI=* go test -bench . -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("allocs", func() {
		Bash(`go test -bench . -benchmem | grep "allocs\|Bench"`, M{"$in": "v1/bench"})
	})

	p.Task("benchjson", func() {
		Bash("go test -bench=BenchmarkLoggerJSON -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("test", func() {
		Run("LOGXI=* go test", M{"$in": "v1"})
		//Run("LOGXI=* go test -run=TestColors", M{"$in": "v1"})
	})

	p.Task("install", func() {
		Run("go get github.com/mattn/go-colorable")
		Run("go get github.com/mattn/go-isatty")
		Run("go get github.com/mgutz/ansi")

		// needed for benchmarks
		Run("go get github.com/Sirupsen/logrus")
		Run("go get gopkg.in/inconshreveable/log15.v2")
	})

	p.Task("demo", func() {
		Run("LOGXI=* go run main.go", M{"$in": "v1/cmd/demo"})
	})
}

func main() {
	Godo(tasks)
}
