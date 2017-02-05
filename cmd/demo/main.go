package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi"
)

var errConfig = fmt.Errorf("file not found")
var dsn = "dbname=testdb"
var logger logxi.Logger
var hostname string
var configFile = "config.json"

func init() {
	hostname, _ = os.Hostname()
}

func loadConfig() {
	logger.Error("Could not read config file", "err", errConfig)
}

func main() {
	// create loggers
	logxi.Trace("creating loggers")
	logger = logxi.New("server")
	modelsLogger := logxi.New("models")

	logger.Debug("Process", "hostname", hostname, "pid", os.Getpid())
	modelsLogger.Info("Connecting to database...")
	modelsLogger.Warn("Could not connect, retrying ...", "dsn", dsn)
	loadConfig()
}
