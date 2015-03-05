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

	logger.Info("Hello, log XI!")

	hostname, _ := os.Hostname()
	logger.Debug("BEGIN main", "hostname", hostname, "pid", os.Getpid())

	logger.Info("Starting server")
	modelsLogger.Info("Connecting to database...")

	logger.Warn("Reconnecting ..", "url", url)

	logger.Error("Could not connect", "tries", 10, "err", errConnection)
}
