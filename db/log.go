package db

import (
	"fmt"
	"log"
	"os"
)

type logger struct {
	logger *log.Logger
	prefix string
}

func newLogger(prefix string) *logger {
	return &logger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		prefix: prefix,
	}
}

func (l *logger) Print(args ...any) {
	l.logger.Print(l.prefix + " " + fmt.Sprint(args...))
}

func (l *logger) Printf(format string, args ...any) {
	l.logger.Printf(l.prefix+" "+format, args...)
}

func (l *logger) Println(args ...any) {
	l.logger.Println(l.prefix, fmt.Sprint(args...))
}

type databaseLogger struct {
	Redis     *logger
	Postgres  *logger
	Typesense *logger
}

var dbLog = &databaseLogger{
	Redis:     newLogger("\033[31m[REDIS]\033[0m"),
	Postgres:  newLogger("\033[38;5;208m[POSTGRES]\033[0m"),
	Typesense: newLogger("\033[34m[TYPESENSE]\033[0m"),
}
