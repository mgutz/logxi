package main

import . "gopkg.in/godo.v1"

func tasks(p *Project) {
	p.Task("bench", func() {
		Run("go test -bench . -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("allocs", func() {
		Bash(`go test -bench . -benchmem | grep "allocs\|Bench"`, M{"$in": "v1/bench"})
	})

	p.Task("benchjson", func() {
		Bash("go test -bench=BenchmarkLoggerJSON -benchmem", M{"$in": "v1/bench"})
	})

	p.Task("test", func() {
		Bash("go test", M{"$in": "v1"})
	})
}

func main() {
	Godo(tasks)
}
