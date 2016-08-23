package log

import (
  "fmt"
)

var TRACE int = 4
var DEBUG int = 3
var INFO int = 2
var WARN int = 1
var ERROR int = 0

var logLevel int = INFO


func SetLevel(level int) {
  logLevel = level
}

func slog(level int, msg string, args ...interface{}) {
  if logLevel >= level {
    fmt.Printf(msg + "\n", args...)
  }
}

func log(level int, msg string) {
  if logLevel >= level {
    fmt.Println(msg)
  }
}

func Strace(msg string, args ...interface{}) {
  slog(TRACE, msg, args...)
}
func Sdebug(msg string, args ...interface{}) {
  slog(DEBUG, msg, args...)
}
func Sinfo(msg string, args ...interface{}) {
  slog(INFO, msg, args...)
}
func Swarn(msg string, args ...interface{}) {
  slog(WARN, msg, args...)
}
func Serror(msg string, args ...interface{}) {
  slog(ERROR, msg, args...)
}

func Trace(msg string) {
  log(TRACE, msg)
}
func Debug(msg string) {
  log(DEBUG, msg)
}
func Info(msg string) {
  log(INFO, msg)
}
func Warn(msg string) {
  log(WARN, msg)
}
func Error(msg string) {
  log(ERROR, msg)
}
