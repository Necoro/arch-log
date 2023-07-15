package log

import (
	"fmt"
	"log"
	"os"
)

var debugLogger = log.New(os.Stdout, "DEBUG: ", 0)
var verboseLogger = log.New(os.Stdout, " INFO: ", 0)
var errorLogger = log.New(os.Stderr, "ERROR: ", 0)
var warnLogger = log.New(os.Stdout, " WARN: ", 0)

type logLevel byte

const (
	debug logLevel = iota
	info
	warn
)

var level logLevel = info

func SetVerbose() {
	level = info
}

func SetDebug() {
	level = debug
}

func IsDebug() bool {
	return level == debug
}

func Debug(v ...any) {
	if level <= debug {
		_ = debugLogger.Output(2, fmt.Sprint(v...))
	}
}

func Debugf(format string, v ...any) {
	if level <= debug {
		_ = debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Print(v ...any) {
	if level <= info {
		_ = verboseLogger.Output(2, fmt.Sprint(v...))
	}
}

func Printf(format string, v ...any) {
	if level <= info {
		_ = verboseLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Error(v ...any) {
	_ = errorLogger.Output(2, fmt.Sprint(v...))
}

// noinspection GoUnusedExportedFunction
func Errorf(format string, a ...any) {
	_ = errorLogger.Output(2, fmt.Sprintf(format, a...))
}

// noinspection GoUnusedExportedFunction
func Warn(v ...any) {
	_ = warnLogger.Output(2, fmt.Sprint(v...))
}

// noinspection GoUnusedExportedFunction
func Warnf(format string, a ...any) {
	_ = warnLogger.Output(2, fmt.Sprintf(format, a...))
}
