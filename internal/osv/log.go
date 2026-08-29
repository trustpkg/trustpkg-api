package osv

import (
	"fmt"
	"log"
	"os"
)

type osvLogger struct {
	logger *log.Logger
	prefix string
}

var osvLog = newOsvLogger("\033[32m[OSV Collector]\033[0m")

func newOsvLogger(prefix string) *osvLogger {
	return &osvLogger{logger: log.New(os.Stderr, "", log.LstdFlags), prefix: prefix}
}

func (logger *osvLogger) Print(args ...any) {
	logger.logger.Print(logger.prefix + " " + fmt.Sprint(args...))
}

func (logger *osvLogger) Printf(format string, args ...any) {
	logger.logger.Printf(logger.prefix+" "+format, args...)
}

func LogError(err error) {
	osvLog.Print(err)
}
