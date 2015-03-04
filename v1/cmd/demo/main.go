package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi/v1"
)

func main() {
	logger := log.New("server")
	modelsLogger := log.New("models")
	logger.Info("Hello, log XI!")

	hostname, _ := os.Hostname()
	logger.Debug("BEGIN main", "hostname", hostname, "pid", os.Getpid())

	logger.Info("Starting server")
	modelsLogger.Info("Connecting to database...")

	url := "http://www.acme.local"
	logger.Warn("Reconnecting ..", "url", url)

	retries := 10
	logger.Error("Could not reconnect", "retries", retries, "err", fmt.Errorf("connection error"))
}
