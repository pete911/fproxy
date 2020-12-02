package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	calldepth = 2
	prefix    = "[INFO]  "
	errPrefix = "[ERROR] "
	logFlags  = log.Ldate | log.Ltime | log.Lshortfile
)

var (
	Verbose      bool
	StdOutLogger = newStdOutLogger(os.Stdout)
	StdErrLogger = newStdErrLogger(os.Stderr)
)

func newStdOutLogger(out io.Writer) *log.Logger {
	return log.New(out, prefix, logFlags)
}

func newStdErrLogger(out io.Writer) *log.Logger {
	return log.New(out, errPrefix, logFlags)
}

func Error(v ...interface{}) {
	StdErrLogger.Output(calldepth, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	StdErrLogger.Output(calldepth, fmt.Sprintf(format, v...))
}

func Log(v ...interface{}) {

	if Verbose {
		StdOutLogger.Output(calldepth, fmt.Sprint(v...))
	}
}

func Logf(format string, v ...interface{}) {

	if Verbose {
		StdOutLogger.Output(calldepth, fmt.Sprintf(format, v...))
	}
}
