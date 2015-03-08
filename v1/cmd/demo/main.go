package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi/v1"
	"github.com/mgutz/logxi/v1/cmd/reldir"
)

var errConnection = fmt.Errorf("connection error")
var url = "http://www.acme.local"
var logger log.Logger
var hostname string

func init() {
	hostname, _ = os.Hostname()
}

func causeError() {
	logger.Error("error in function", "err", errConnection)
}

func main() {
	// create the loggers
	log.Trace("creating loggers")
	logger = log.New("server")
	modelsLogger := log.New("models")

	logger.Debug("Process", "hostname", hostname, "pid", os.Getpid())
	logger.Info("Starting server...")
	reldir.Foo()

	causeError()

	modelsLogger.Info("Connecting to database...")
	logger.Warn("Reconnecting ..", "url", url)
}
