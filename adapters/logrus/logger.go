package logrus

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/tcncloud/wollemi/ports/logging"
)

func NewLogger(out io.Writer) *Logger {
	log := &logrus.Logger{
		Formatter: &logrus.TextFormatter{},
		Level:     logrus.InfoLevel,
		ExitFunc:  os.Exit,
		Out:       out,
	}

	return &Logger{
		entry:  logrus.NewEntry(log),
		logrus: log,
	}
}

type Logger struct {
	logrus *logrus.Logger
	entry  *logrus.Entry
}

func (this *Logger) WithError(err error) logging.Logger {
	return &Logger{
		entry:  this.entry.WithError(err),
		logrus: this.logrus,
	}
}

func (this *Logger) WithFields(fields logging.Fields) logging.Logger {
	return &Logger{
		entry:  this.entry.WithFields(logrus.Fields(fields)),
		logrus: this.logrus,
	}
}

func (this *Logger) WithField(keys string, value interface{}) logging.Logger {
	return &Logger{
		entry:  this.entry.WithField(keys, value),
		logrus: this.logrus,
	}
}

func (this *Logger) Infof(format string, args ...interface{}) {
	this.entry.Infof(format, args...)
}

func (this *Logger) Info(args ...interface{}) {
	this.entry.Info(args...)
}

func (this *Logger) Warnf(format string, args ...interface{}) {
	this.entry.Warnf(format, args...)
}

func (this *Logger) Warn(args ...interface{}) {
	this.entry.Warn(args...)
}

func (this *Logger) Error(args ...interface{}) {
	this.entry.Error(args...)
}

func (this *Logger) Debug(args ...interface{}) {
	this.entry.Debug(args...)
}

func (this *Logger) GetLevel() logging.Level {
	return logging.Level(this.logrus.GetLevel())
}

func (this *Logger) SetLevel(lvl logging.Level) {
	switch lvl {
	case logging.PanicLevel:
		this.logrus.SetLevel(logrus.PanicLevel)
	case logging.FatalLevel:
		this.logrus.SetLevel(logrus.FatalLevel)
	case logging.ErrorLevel:
		this.logrus.SetLevel(logrus.ErrorLevel)
	case logging.WarnLevel:
		this.logrus.SetLevel(logrus.WarnLevel)
	case logging.InfoLevel:
		this.logrus.SetLevel(logrus.InfoLevel)
	case logging.DebugLevel:
		this.logrus.SetLevel(logrus.DebugLevel)
	case logging.TraceLevel:
		this.logrus.SetLevel(logrus.TraceLevel)
	}
}

func (*Logger) Exit(code int) {
	logrus.Exit(code)
}

func (this *Logger) SetFormatter(formatter logging.Formatter) {
	switch formatter.(type) {
	case *logging.TextFormatter:
		this.logrus.SetFormatter(&logrus.TextFormatter{})
	case *logging.JsonFormatter:
		this.logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
