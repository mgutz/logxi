package main

import (
	"fmt"

	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func main() {
	logger = log.New("logxi")
	logger.Debug("I'm debugging", "fruit", "apple", "balance", 42.0)
	logger.Info("Psst. Can you keep a secret?")
	logger.Warn("Hmm ...", "balance", 0.0)
	logger.Error("Oh oh", "err", fmt.Errorf("some error"), "balance", -1000.0)
}
