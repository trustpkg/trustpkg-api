package typesense

import (
	"fmt"
	"log"
	"os"
)

type typesenseLogger struct {
	logger *log.Logger
	prefix string
}

var typesenseLog = newTypesenseLogger("\033[31m[TYPESENSE]\033[0m")

func newTypesenseLogger(prefix string) *typesenseLogger {
	return &typesenseLogger{logger: log.New(os.Stderr, "", log.LstdFlags), prefix: prefix}
}

func (logger *typesenseLogger) Print(args ...any) {
	logger.logger.Print(logger.prefix + " " + fmt.Sprint(args...))
}

func (logger *typesenseLogger) Printf(format string, args ...any) {
	logger.logger.Printf(logger.prefix+" "+format, args...)
}

func LogError(err error) {
	typesenseLog.Print(err)
}
