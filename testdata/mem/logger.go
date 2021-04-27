package mem

import (
	"fmt"
	"sync"
	"time"

	"github.com/tcncloud/wollemi/ports/logging"
)

func NewLogger() *Logger {
	return &Logger{lvl: logging.TraceLevel}
}

type Logger struct {
	lines []map[string]interface{}
	lvl   logging.Level
	mu    sync.Mutex
}

func (this *Logger) Lines() []map[string]interface{} {
	return this.lines
}

func (this *Logger) Exit(int) {}

func (this *Logger) GetLevel() logging.Level {
	return this.lvl
}

func (this *Logger) SetLevel(lvl logging.Level) {
	switch lvl {
	case logging.PanicLevel:
		this.lvl = logging.PanicLevel
	case logging.FatalLevel:
		this.lvl = logging.FatalLevel
	case logging.ErrorLevel:
		this.lvl = logging.ErrorLevel
	case logging.WarnLevel:
		this.lvl = logging.WarnLevel
	case logging.InfoLevel:
		this.lvl = logging.InfoLevel
	case logging.DebugLevel:
		this.lvl = logging.DebugLevel
	case logging.TraceLevel:
		this.lvl = logging.TraceLevel
	}
}

func (this *Logger) SetFormatter(logging.Formatter) {}

func (this *Logger) Infof(format string, args ...interface{}) {
	this.Logf("info", nil, format, args...)
}

func (this *Logger) Warnf(format string, args ...interface{}) {
	this.Logf("warn", nil, format, args...)
}

func (this *Logger) Info(args ...interface{}) {
	this.Log("info", nil, args...)
}

func (this *Logger) Debugf(format string, args ...interface{}) {
	this.Logf("debug", nil, format, args...)
}

func (this *Logger) Debug(args ...interface{}) {
	this.Log("debug", nil, args...)
}

func (this *Logger) Warn(args ...interface{}) {
	this.Log("warn", nil, args...)
}

func (this *Logger) Error(args ...interface{}) {
	this.Log("error", nil, args...)
}

func (this *Logger) Logf(lvl string, entry map[string]interface{}, format string, args ...interface{}) {
	this.Log(lvl, entry, fmt.Sprintf(format, args...))
}

func (this *Logger) Log(lvl string, entry map[string]interface{}, args ...interface{}) {
	level, err := logging.ParseLevel(lvl)
	if err != nil {
		panic(err)
	}

	if level > this.GetLevel() {
		return
	}

	if entry == nil {
		entry = make(map[string]interface{})
	}

	entry["time"] = time.Now()
	entry["level"] = lvl
	entry["msg"] = fmt.Sprint(args...)

	this.mu.Lock()
	this.lines = append(this.lines, entry)
	this.mu.Unlock()
}

func (this *Logger) WithFields(fields logging.Fields) logging.Logger {
	var out logging.Logger = this

	for name, value := range fields {
		out = out.WithField(name, value)
	}

	return out
}

func (this *Logger) WithError(err error) logging.Logger {
	return this.WithField("error", err)
}

func (this *Logger) WithField(name string, value interface{}) logging.Logger {
	return &LoggerEntry{
		entry:  map[string]interface{}{name: value},
		logger: this,
	}
}

type LoggerEntry struct {
	logger *Logger
	entry  map[string]interface{}
}

func (*LoggerEntry) SetFormatter(logging.Formatter) {}

func (*LoggerEntry) Exit(int) {}

func (this *LoggerEntry) SetLevel(lvl logging.Level) {
	this.logger.SetLevel(lvl)
}

func (this *LoggerEntry) GetLevel() logging.Level {
	return this.logger.GetLevel()
}

func (this *LoggerEntry) WithError(err error) logging.Logger {
	return this.WithField("error", err)
}

func (this *LoggerEntry) WithFields(fields logging.Fields) logging.Logger {
	var out logging.Logger = this

	for name, value := range fields {
		out = out.WithField(name, value)
	}

	return out
}

func (this *LoggerEntry) WithField(name string, value interface{}) logging.Logger {
	out := &LoggerEntry{
		entry:  make(map[string]interface{}),
		logger: this.logger,
	}

	for name, value := range this.entry {
		out.entry[name] = value
	}

	out.entry[name] = value

	return out
}

func (this *LoggerEntry) Logf(lvl, format string, args ...interface{}) {
	this.Log(lvl, fmt.Sprintf(format, args...))
}

func (this *LoggerEntry) Log(lvl string, args ...interface{}) {
	entry := make(map[string]interface{})

	for name, value := range this.entry {
		entry[name] = value
	}

	this.logger.Log(lvl, entry, args...)
}

func (this *LoggerEntry) Infof(format string, args ...interface{}) {
	this.Logf("info", format, args...)
}

func (this *LoggerEntry) Warnf(format string, args ...interface{}) {
	this.Logf("warn", format, args...)
}

func (this *LoggerEntry) Info(args ...interface{}) {
	this.Log("info", args...)
}

func (this *LoggerEntry) Debugf(format string, args ...interface{}) {
	this.Logf("debug", format, args...)
}

func (this *LoggerEntry) Debug(args ...interface{}) {
	this.Log("debug", args...)
}

func (this *LoggerEntry) Warn(args ...interface{}) {
	this.Log("warn", args...)
}

func (this *LoggerEntry) Error(args ...interface{}) {
	this.Log("error", args...)
}
