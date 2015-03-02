package main

import (
	"github.com/mgutz/logxi/v1"
)

func main() {
	log.Info("This is a test app")
	l := log.NewColorable("bench")
	l.SetLevel(log.LevelDebug)
	l.SetFormatter(log.NewHappyDevFormatter("bench"))
	l.Debug("just another day", "key")
	l.Debug("and another one", "key")
	l.Info("something you should know")
	l.Warn("hmm didn't expect that")
	l.Error("oh oh, you're in trouble", "key", 1)
}
