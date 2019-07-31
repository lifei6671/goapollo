package goapollo

import (
	"log"
	"os"
)

var logger ILogger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

type ILogger interface {
	Printf(format string, args ...interface{})
}

//SetILogger 设置日志接口.
func SetILogger(logger1 ILogger) {
	logger = logger1
}
