package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func main() {
	logger = log.New("mylogger")
	logger.Info("Hello, log XI!")

	hostname, _ := os.Hostname()
	logger.Debug("BEGIN main", "hostname", hostname, "pid", os.Getpid())

	logger.Info("Starting server")

	url := "http://www.acme.local"
	logger.Warn("Reconnecting ..", "url", url)

	retries := 10
	logger.Error("Could not reconnect", "retries", retries, "err", fmt.Errorf("connection error"))
}
