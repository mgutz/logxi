package bench

import (
	"encoding/json"
	"io/ioutil"
	L "log"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/mgutz/logxi"
	"github.com/uber-go/zap"
	"gopkg.in/inconshreveable/log15.v2"
)

type M map[string]interface{}

var testObject = M{
	"foo": "bar",
	"bah": M{
		"int":      1,
		"float":    -100.23,
		"date":     "06-01-01T15:04:05-0700",
		"bool":     true,
		"nullable": nil,
	},
}

var pid = os.Getpid()

var writer = ioutil.Discard

func toJSON(m map[string]interface{}) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// These tests write out all log levels with concurrency turned on and
// equivalent fields.

func BenchmarkLog(b *testing.B) {
	//fmt.Println("")
	l := L.New(writer, "bench ", L.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debug := map[string]interface{}{"l": "debug", "key1": 1, "key2": "string", "key3": false}
		l.Printf(toJSON(debug))

		info := map[string]interface{}{"l": "info", "key1": 1, "key2": "string", "key3": false}
		l.Printf(toJSON(info))

		warn := map[string]interface{}{"l": "warn", "key1": 1, "key2": "string", "key3": false}
		l.Printf(toJSON(warn))

		err := map[string]interface{}{"l": "error", "key1": 1, "key2": "string", "key3": false}
		l.Printf(toJSON(err))
	}
	b.StopTimer()
}

func BenchmarkLogComplex(b *testing.B) {
	//fmt.Println("")
	l := L.New(writer, "bench ", L.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debug := map[string]interface{}{"l": "debug", "key1": 1, "obj": testObject}
		l.Printf(toJSON(debug))

		info := map[string]interface{}{"l": "info", "key1": 1, "obj": testObject}
		l.Printf(toJSON(info))

		warn := map[string]interface{}{"l": "warn", "key1": 1, "obj": testObject}
		l.Printf(toJSON(warn))

		err := map[string]interface{}{"l": "error", "key1": 1, "obj": testObject}
		l.Printf(toJSON(err))
	}
	b.StopTimer()
}

func BenchmarkLogxi(b *testing.B) {
	stdout := logxi.NewConcurrentWriter(writer)
	l := logxi.NewLogger3(stdout, "bench", logxi.NewJSONFormatter("bench"))
	l.SetLevel(logxi.LevelDebug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "key2", "string", "key3", false)
		l.Info("info", "key", 1, "key2", "string", "key3", false)
		l.Warn("warn", "key", 1, "key2", "string", "key3", false)
		l.Error("error", "key", 1, "key2", "string", "key3", false)
	}
	b.StopTimer()
}

func BenchmarkLogxiComplex(b *testing.B) {
	//fmt.Println("")
	stdout := logxi.NewConcurrentWriter(writer)
	l := logxi.NewLogger3(stdout, "bench", logxi.NewJSONFormatter("bench"))
	l.SetLevel(logxi.LevelDebug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "obj", testObject)
		l.Info("info", "key", 1, "obj", testObject)
		l.Warn("warn", "key", 1, "obj", testObject)
		l.Error("error", "key", 1, "obj", testObject)
	}
	b.StopTimer()

}

func BenchmarkLogrus(b *testing.B) {
	//fmt.Println("")
	l := logrus.New()
	l.Out = writer
	l.Formatter = &logrus.JSONFormatter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "key2": "string", "key3": false}).Debug("debug")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "key2": "string", "key3": false}).Info("info")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "key2": "string", "key3": false}).Warn("warn")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "key2": "string", "key3": false}).Error("error")
	}
	b.StopTimer()
}

func BenchmarkLogrusComplex(b *testing.B) {
	//fmt.Println("")
	l := logrus.New()
	l.Out = writer
	l.Formatter = &logrus.JSONFormatter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "obj": testObject}).Debug("debug")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "obj": testObject}).Info("info")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "obj": testObject}).Warn("warn")
		l.WithFields(logrus.Fields{"_n": "bench", "_p": pid, "key": 1, "obj": testObject}).Error("error")
	}
	b.StopTimer()
}

func BenchmarkLog15(b *testing.B) {
	//fmt.Println("")
	l := log15.New(log15.Ctx{"_n": "bench", "_p": pid})
	l.SetHandler(log15.SyncHandler(log15.StreamHandler(writer, log15.JsonFormat())))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "key2", "string", "key3", false)
		l.Info("info", "key", 1, "key2", "string", "key3", false)
		l.Warn("warn", "key", 1, "key2", "string", "key3", false)
		l.Error("error", "key", 1, "key2", "string", "key3", false)
	}
	b.StopTimer()

}

func BenchmarkLog15Complex(b *testing.B) {
	//fmt.Println("")
	l := log15.New(log15.Ctx{"_n": "bench", "_p": pid})
	l.SetHandler(log15.SyncHandler(log15.StreamHandler(writer, log15.JsonFormat())))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "obj", testObject)
		l.Info("info", "key", 1, "obj", testObject)
		l.Warn("warn", "key", 1, "obj", testObject)
		l.Error("error", "key", 1, "obj", testObject)
	}
	b.StopTimer()
}

func BenchmarkZap(b *testing.B) {
	l := zap.New(
		zap.NewJSONEncoder(),
		zap.DiscardOutput,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", zap.Int64("key", 1), zap.String("key2", "string"), zap.Bool("key3", false))
		l.Info("info", zap.Int64("key", 1), zap.String("key2", "string"), zap.Bool("key3", false))
		l.Warn("warn", zap.Int64("key", 1), zap.String("key2", "string"), zap.Bool("key3", false))
		l.Error("error", zap.Int64("key", 1), zap.String("key2", "string"), zap.Bool("key3", false))
	}
	b.StopTimer()
}

func BenchmarkZapComplex(b *testing.B) {
	l := zap.New(
		zap.NewJSONEncoder(),
		zap.DiscardOutput,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", zap.Int64("key", 1), zap.Object("obj", testObject))
		l.Info("info", zap.Int64("key", 1), zap.Object("obj", testObject))
		l.Warn("warn", zap.Int64("key", 1), zap.Object("obj", testObject))
		l.Error("error", zap.Int64("key", 1), zap.Object("obj", testObject))
	}
	b.StopTimer()
}
