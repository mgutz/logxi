package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi/v1"
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
	logger = log.New("server")
	modelsLogger := log.New("models")

	log.Debug("I'm the default logger")

	logger.Info("BEGIN main", "hostname", hostname, "pid", os.Getpid())

	causeError()

	modelsLogger.Info("Connecting to database...")
	logger.Warn("Reconnecting ..", "url", url)
}
