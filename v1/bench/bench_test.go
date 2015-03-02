package bench

import (
	"fmt"
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
		"date":     "testObject006-01-0testObjectT15:04:05-0700",
		"bool":     true,
		"nullable": nil,
	},
}

var simpleArgs = []interface{}{"key", 1, "key2", "string", "key3", false}
var complexArgs = []interface{}{"key", 1, "obj", testObject}

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
	l := log.NewLogger(os.Stdout, "bench")
	l.SetLevel(log.LevelDebug)
	l.SetFormatter(log.NewJSONFormatter("bench"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", simpleArgs...)
		l.Info("info", simpleArgs...)
		l.Warn("warn", simpleArgs...)
		l.Error("error", simpleArgs...)
	}
	b.StopTimer()
	log.DisableColors(false)
}

func BenchmarkLogxiComplex(b *testing.B) {
	l := log.NewLogger(os.Stdout, "bench")
	l.SetLevel(log.LevelDebug)
	l.SetFormatter(log.NewJSONFormatter("bench"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", complexArgs...)
		l.Info("info", complexArgs...)
		l.Warn("warn", complexArgs...)
		l.Error("error", complexArgs...)
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
		l.Debug("debug", simpleArgs...)
		l.Info("info", simpleArgs...)
		l.Warn("warn", simpleArgs...)
		l.Error("error", simpleArgs...)
	}
	b.StopTimer()

}

func BenchmarkLog15Complex(b *testing.B) {
	l := log15.New(log15.Ctx{"m": "bench"})
	l.SetHandler(log15.StreamHandler(os.Stdout, log15.JsonFormat()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("debug", complexArgs...)
		l.Info("info", complexArgs...)
		l.Warn("warn", complexArgs...)
		l.Error("error", complexArgs...)
	}
	b.StopTimer()

}

func causeError() error {
	return fmt.Errorf("here is an error")
}

func nestedError() error {
	return causeError()
}
