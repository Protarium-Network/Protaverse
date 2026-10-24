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

func LogInfo(msg string)    { logMessage("INFO", ColorInfo, msg) }
func LogWarn(msg string)    { logMessage("WARN", ColorWarn, msg) }
func LogError(msg string)   { logMessage("ERROR", ColorError, msg) }
func LogSuccess(msg string) { logMessage("SUCCESS", ColorSuccess, msg) }
