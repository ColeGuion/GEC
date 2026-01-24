package print

import (
	"fmt"
	"io"
	"os"
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

// Critical prints critical messages
func Critical(format string, args ...interface{}) {
	if shouldPrint(LevelCritical) {
		msg := fmt.Sprintf(format, args...)
		printCritical(msg)
	}
}

// Error prints error messages
func Error(format string, args ...interface{}) {
	if shouldPrint(LevelError) {
		msg := fmt.Sprintf(format, args...)
		printError(msg)
	}
}

// Warning prints warning messages
func Warning(format string, args ...interface{}) {
	if shouldPrint(LevelWarning) {
		msg := fmt.Sprintf(format, args...)
		printMessage("WARNING", msg)
	}
}

// Info prints info messages
func Info(format string, args ...interface{}) {
	if shouldPrint(LevelInfo) {
		msg := fmt.Sprintf(format, args...)
		printMessage("INFO", msg)
	}
}

// Debug prints debug messages
func Debug(format string, args ...interface{}) {
	if shouldPrint(LevelDebug) {
		msg := fmt.Sprintf(format, args...)
		printMessage("DEBUG", msg)
	}
}

// printCritical prints a critical message with red background and white text
func printCritical(message string) {
	// ANSI escape codes for red background and white text
	const criticalColor = "\033[1;37;41m" // White text on red background
	const resetColor = "\033[0m"

	fmt.Printf("%s[CRITICAL] %s%s\n", criticalColor, message, resetColor)
}

// printError prints an error message in red color
func printError(message string) {
	// ANSI escape codes for red color
	const redColor = "\033[1;31m"
	const resetColor = "\033[0m"

	fmt.Printf("%s[ERROR] %s%s\n", redColor, message, resetColor)
}

// printMessage prints a formatted message with timestamp and level
func printMessage(level, message string) {
	//fmt.Printf("%s\n", message)
	fmt.Printf("[%s] %s\n", level, message)
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