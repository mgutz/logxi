package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi/v1"
)

var errConnection = fmt.Errorf("connection error")
var url = "http://www.acme.local"

func main() {
	// create the loggers
	logger := log.New("server")
	modelsLogger := log.New("models")

	log.Info("I'm the default logger")

	logger.Info("Hello, log XI!")

	hostname, _ := os.Hostname()
	logger.Debug("BEGIN main", "hostname", hostname, "pid", os.Getpid())
	logger.Info("Starting server")

	logger.Error("Could not connect", "err", errConnection)
	modelsLogger.Info("Connecting to database...")
	logger.Warn("Reconnecting ..", "url", url)
}
