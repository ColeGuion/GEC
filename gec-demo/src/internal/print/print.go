package print

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Log levels
const (
	LevelCritical = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
	LevelDisabled
)

var (
	// Current log level - default to INFO
	logLevel = LevelInfo
	mu       sync.RWMutex
)

// SetLevel sets the current log level
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = level
}

// GetLevel returns the current log level
func GetLevel() int {
	mu.RLock()
	defer mu.RUnlock()
	return logLevel
}

// shouldPrint checks if a message at the given level should be printed
func shouldPrint(level int) bool {
	mu.RLock()
	defer mu.RUnlock()
	return level <= logLevel && logLevel != LevelDisabled
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo() (string, int) {
	// Get the caller's file and line number (skip 2 frames to get the actual caller)
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	
	// Get just the filename without the full path
	return filepath.Base(file), line
}

// Critical prints critical messages
func Critical(format string, args ...interface{}) {
	if shouldPrint(LevelCritical) {
		file, line := getCallerInfo()
		msg := fmt.Sprintf(format, args...)
		printCritical(file, line, msg)
	}
}

// Error prints error messages
func Error(format string, args ...interface{}) {
	if shouldPrint(LevelError) {
		file, line := getCallerInfo()
		msg := fmt.Sprintf(format, args...)
		printError(file, line, msg)
	}
}

// Warning prints warning messages
func Warning(format string, args ...interface{}) {
	if shouldPrint(LevelWarning) {
		file, line := getCallerInfo()
		msg := fmt.Sprintf(format, args...)
		printWarning(file, line, msg)
	}
}

// Info prints info messages
func Info(format string, args ...interface{}) {
	if shouldPrint(LevelInfo) {
		file, line := getCallerInfo()
		msg := fmt.Sprintf(format, args...)
		printMessage("INFO", file, line, msg)
	}
}

// Debug prints debug messages
func Debug(format string, args ...interface{}) {
	if shouldPrint(LevelDebug) {
		file, line := getCallerInfo()
		msg := fmt.Sprintf(format, args...)
		printMessage("DEBUG", file, line, msg)
	}
}

func printCritical(file string, line int, message string) {
	const criticalColor = "\x1b[1;37;41m" // White text on red background
	const resetColor = "\x1b[0m"
	fmt.Printf("%s:%d: %s[CRITICAL] %s%s\n", file, line, criticalColor, message, resetColor)
}

func printError(file string, line int, message string) {
	const redColor = "\x1b[91m"
	const resetColor = "\x1b[0m"
	fmt.Printf("%s:%d: %s[ERROR] %s%s\n", file, line, redColor, message, resetColor)
}

func printWarning(file string, line int, message string) {
	const warnColor = "\x1b[91m"
	const resetColor = "\x1b[0m"
	fmt.Printf("%s:%d: %sWARNING:%s %s\n", file, line, warnColor, resetColor, message)
}

// printMessage prints a formatted message with timestamp and level
func printMessage(level, file string, line int, message string) {
	fmt.Printf("%s:%d: %s\n", file, line, message)
	//fmt.Printf("[%s] %s:%d: %s\n", level, file, line, message)
}

// Helper functions to set specific log levels
func SetLevelCritical() { SetLevel(LevelCritical) }
func SetLevelError()    { SetLevel(LevelError) }
func SetLevelWarning()  { SetLevel(LevelWarning) }
func SetLevelInfo()     { SetLevel(LevelInfo) }
func SetLevelDebug()    { SetLevel(LevelDebug) }
func DisableLogging()   { SetLevel(LevelDisabled) }

// SetOutput allows redirecting output (optional feature)
var output io.Writer = os.Stdout

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}