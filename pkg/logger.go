package pkg

import (
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/go-kit/kit/log"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	kitzap "github.com/go-kit/kit/log/zap"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger log.Logger
var _loggerOnce sync.Once

func NewDefaultLogger() {
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
}

func NewLogrusLogger() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:03:04",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File)
			funcName := path.Base(frame.Function)
			return funcName, fileName
		},
	})
	logger = kitlogrus.NewLogger(logrus.New())
}

func NewZapLogger() {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	logger = kitzap.NewZapSugarLogger(zap.New(core), zapcore.DebugLevel)
}

func GetLogger() log.Logger {
	_loggerOnce.Do(func() {
		NewDefaultLogger()
	})
	return logger
}
