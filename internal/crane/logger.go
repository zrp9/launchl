// Package crane provides a logger
package crane

import (
	"fmt"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type LogLevel int

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
	Fatal
	Panic
)

const (
	LogPrefix = "var/log"
	LogName   = "logs.log"
)

type Zfields map[string]any

func (l LogLevel) String() string {
	return [...]string{"Trace", "Debug", "Info", "Warn", "Error", "Fatal", "Panic"}[l]
}

func (l LogLevel) Index() int {
	return int(l)
}

var logfile = NewLogFile(
	WithFilename(fmt.Sprintf("%s/%s", LogPrefix, LogName)),
)

var DefaultLogger = NewLogger(
	logfile,
	WithLevel("trace"),
)

func defaultLogFile() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("%v/%v", LogPrefix, LogName),
		MaxSize:    15,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}
}

func NewLogFile(ops ...func(*lumberjack.Logger)) *lumberjack.Logger {
	logfile := defaultLogFile()
	for _, op := range ops {
		op(logfile)
	}
	return logfile
}

func WithFilename(n string) func(*lumberjack.Logger) {
	return func(l *lumberjack.Logger) {
		l.Filename = n
	}
}

func WithMaxSize(s int) func(*lumberjack.Logger) {
	return func(l *lumberjack.Logger) {
		l.MaxSize = s
	}
}

func WithMaxBackups(s int) func(*lumberjack.Logger) {
	return func(l *lumberjack.Logger) {
		l.MaxBackups = s
	}
}

func WithMaxAge(a int) func(*lumberjack.Logger) {
	return func(l *lumberjack.Logger) {
		l.MaxAge = a
	}
}

func WithCompress(b bool) func(*lumberjack.Logger) {
	return func(l *lumberjack.Logger) {
		l.Compress = b
	}
}

type Zlogrus struct {
	Logfile *lumberjack.Logger
	logger  *logrus.Logger
}

func (z Zlogrus) MustTrace(msg string) {
	z.logger.Trace(msg)
}

func (z Zlogrus) MustDebug(msg string) {
	z.logger.Debug(msg)
}

func (z Zlogrus) MustDebugErr(err error) {
	z.logger.Debug(err.Error())
}

func (z Zlogrus) MustInfo(msg string) {
	z.logger.Info(msg)
}

func (z Zlogrus) MustWarn(msg string) {
	z.logger.Warn(msg)
}

func (z Zlogrus) MustError(e error) {
	z.logger.Error(e)
}

func (z Zlogrus) MustFatal(msg string) {
	z.logger.Fatal(msg)
}

func (z Zlogrus) MustPanic(msg string) {
	z.logger.Panic(msg)
}

func (z Zlogrus) LogFields(fields logrus.Fields) {
	z.logger.WithFields(fields).Debug()
}

func (z Zlogrus) Zlog(fields Zfields) {
	lfs := logrus.Fields{}
	for fieldName, value := range fields {
		if _, exists := lfs[fieldName]; !exists {
			lfs[fieldName] = value
		}
	}
	z.logger.WithFields(lfs)
}

func initLogger() *Zlogrus {
	return &Zlogrus{
		logger: logrus.New(),
	}
}

func NewLogger(logfile *lumberjack.Logger, ops ...func(*Zlogrus)) *Zlogrus {
	zlog := initLogger()
	zlog.logger.SetOutput(logfile)
	for _, op := range ops {
		op(zlog)
	}
	return zlog
}

func WithJSONFormatter(b bool) func(*Zlogrus) {
	return func(z *Zlogrus) {
		if b {
			z.logger.SetFormatter(&logrus.JSONFormatter{
				DisableTimestamp: false,
				TimestampFormat:  time.Kitchen,
			})
		}
	}
}

func WitTextFormatter(b bool) func(*Zlogrus) {
	return func(z *Zlogrus) {
		if b {
			z.logger.SetFormatter(&logrus.TextFormatter{
				DisableColors:    true,
				FullTimestamp:    true,
				TimestampFormat:  time.Layout,
				PadLevelText:     true,
				QuoteEmptyFields: true,
			})
		}
	}
}

func WithLevel(lvl string) func(*Zlogrus) {
	return func(z *Zlogrus) {
		var loglevel logrus.Level
		switch strings.ToLower(lvl) {
		case "trace":
			loglevel = logrus.TraceLevel
		case "debug":
			loglevel = logrus.DebugLevel
		case "info":
			loglevel = logrus.InfoLevel
		case "warn":
			loglevel = logrus.WarnLevel
		case "error":
			loglevel = logrus.ErrorLevel
		case "fatal":
			loglevel = logrus.FatalLevel
		case "panic":
			loglevel = logrus.PanicLevel
		default:
			loglevel = logrus.TraceLevel
		}
		z.logger.SetLevel(loglevel)
	}
}

func WithFields(args map[string]any) func(*Zlogrus) {
	return func(z *Zlogrus) {
		for key, val := range args {
			z.logger.WithField(key, val)
		}
	}
}

func WithReportCaller(b bool) func(*Zlogrus) {
	return func(z *Zlogrus) {
		z.logger.SetReportCaller(b)
	}
}
