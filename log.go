package goapollo

import (
	"log"
	"os"
)

var logger = NewLocalLogger()

type ILogger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type localLogger struct {
	l *log.Logger
}

func NewLocalLogger() ILogger {
	return &localLogger{
		l: log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags),
	}
}
func (l *localLogger) Errorf(format string, args ...interface{}) {
	l.l.Printf(format, args...)
}

func (l *localLogger) Infof(format string, args ...interface{}) {
	l.l.Printf(format, args...)
}

//SetILogger 设置日志接口.
func SetILogger(logger1 ILogger) {
	logger = logger1
}
