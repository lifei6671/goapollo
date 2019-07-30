package goapollo

import (
	"log"
	"os"
)

var logger = NewLocalLogger()

type ILogger interface {
	Printf(format string, args ...interface{})
}

type localLogger struct {
	*log.Logger
}

func NewLocalLogger() ILogger {
	return &localLogger{
		log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags),
	}
}

//SetILogger 设置日志接口.
func SetILogger(logger1 ILogger) {
	logger = logger1
}
