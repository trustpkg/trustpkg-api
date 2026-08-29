package npm

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	npmCollectorLog  = newNpmLogger("\033[33m[NPM COLLECTOR]\033[0m")
	npmReplicatorLog = newNpmLogger("\033[35m[NPM_REPLICATOR]\033[0m")
)

type npmLogger struct {
	logger *log.Logger
	prefix string
}

func newNpmLogger(prefix string) *npmLogger {
	return &npmLogger{logger: log.New(os.Stderr, "", log.LstdFlags), prefix: prefix}
}

func (logger *npmLogger) Print(args ...any) {
	logger.logger.Print(logger.prefix + " " + fmt.Sprint(args...))
}

func (logger *npmLogger) Printf(format string, args ...any) {
	logger.logger.Printf(logger.prefix+" "+format, args...)
}

func (logger *npmLogger) Println(args ...any) {
	logger.logger.Print(logger.prefix + " " + strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func LogCollectorError(err error) {
	npmCollectorLog.Print(err)
}

func LogReplicatorError(err error) {
	npmReplicatorLog.Print(err)
}
