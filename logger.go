package main

import (
	"fmt"
	"time"
)

const (
	ColorReset   = "\x1b[0m"
	ColorInfo    = "\x1b[36m"
	ColorWarn    = "\x1b[33m"
	ColorError   = "\x1b[31m"
	ColorSuccess = "\x1b[32m"
)

func logMessage(level, color, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s[%s] [%s] %s%s\n", color, timestamp, level, message, ColorReset)
}

func LogInfo(format string, a ...any)  { logMessage("INFO", ColorInfo, fmt.Sprintf(format, a...)) }
func LogWarn(format string, a ...any)  { logMessage("WARN", ColorWarn, fmt.Sprintf(format, a...)) }
func LogError(format string, a ...any) { logMessage("ERROR", ColorError, fmt.Sprintf(format, a...)) }
func LogSuccess(format string, a ...any) {
	logMessage("SUCCESS", ColorSuccess, fmt.Sprintf(format, a...))
}
