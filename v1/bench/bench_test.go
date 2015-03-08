package bench

import (
	L "log"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/mgutz/logxi/v1"
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

func BenchmarkBuiltinLog(b *testing.B) {
	l := L.New(os.Stdout, " [log] ", L.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Printf("debug %s=%v %s=%v \n", "key", 1, "other", testObject)
		l.Printf("info %s=%v %s=%v \n", "key", 1, "other", testObject)
		l.Printf("warn %s=%v %s=%v \n", "key", 1, "other", testObject)
		l.Printf("error %s=%v %s=%v \n", "key", 1, "other", testObject)
	}
	b.StopTimer()
}

func BenchmarkLogxi(b *testing.B) {
	log.DisableColors(true)
	l := log.NewLogger3(os.Stdout, "bench", log.NewJSONFormatter("bench"))
	l.SetLevel(log.LevelDebug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "key2", "string", "key3", false)
		l.Info("info", "key", 1, "key2", "string", "key3", false)
		l.Warn("warn", "key", 1, "key2", "string", "key3", false)
		l.Error("error", "key", 1, "key2", "string", "key3", false)
	}
	b.StopTimer()
	log.DisableColors(false)
}

func BenchmarkLogxiComplex(b *testing.B) {
	l := log.NewLogger3(os.Stdout, "bench", log.NewJSONFormatter("bench"))
	l.SetLevel(log.LevelDebug)

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
	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "key2": "string", "key3": false}).Debug("debug")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "key2": "string", "key3": false}).Info("info")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "key2": "string", "key3": false}).Warn("warn")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "key2": "string", "key3": false}).Error("error")
	}
	b.StopTimer()
}

func BenchmarkLogrusComplex(b *testing.B) {
	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "obj": testObject}).Debug("debug")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "obj": testObject}).Info("info")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "obj": testObject}).Warn("warn")
		l.WithFields(logrus.Fields{"m": "bench", "key": 1, "obj": testObject}).Error("error")
	}
	b.StopTimer()
}

func BenchmarkLog15(b *testing.B) {
	l := log15.New(log15.Ctx{"m": "bench"})
	l.SetHandler(log15.StreamHandler(os.Stdout, log15.JsonFormat()))

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
	l := log15.New(log15.Ctx{"m": "bench"})
	l.SetHandler(log15.StreamHandler(os.Stdout, log15.JsonFormat()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", "key", 1, "obj", testObject)
		l.Info("info", "key", 1, "obj", testObject)
		l.Warn("warn", "key", 1, "obj", testObject)
		l.Error("error", "key", 1, "obj", testObject)
	}
	b.StopTimer()
}
