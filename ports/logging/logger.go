package logging

import (
	"fmt"
	"strings"
)

type Logger interface {
	WithError(error) Logger
	WithFields(Fields) Logger
	WithField(string, interface{}) Logger
	Infof(string, ...interface{})
	Info(...interface{})
	Warnf(string, ...interface{})
	Warn(...interface{})
	Error(...interface{})
	Debug(...interface{})
	GetLevel() Level
	SetLevel(Level)
	SetFormatter(Formatter)
	Exit(int)
}

type Fields map[string]interface{}

type Level uint32

func (lvl Level) String() string {
	switch lvl {
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warn"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case TraceLevel:
		return "trace"
	default:
		return "invalid"
	}
}

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid level: %q", lvl)
}

func ParseFormat(format string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "json":
		return &JsonFormatter{}, nil
	case "text":
		return &TextFormatter{}, nil
	}

	return nil, fmt.Errorf("not a valid format: %q", format)
}

type Formatter interface{ isFormatter() }

type JsonFormatter struct{}
type TextFormatter struct{}

func (*JsonFormatter) isFormatter() {}
func (*TextFormatter) isFormatter() {}
