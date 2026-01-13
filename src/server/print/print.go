package print

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
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
	return level >= logLevel
}

// Debug prints debug messages (visible at DEBUG level and above)
func Debug(format string, args ...interface{}) {
	if shouldPrint(LevelDebug) {
		msg := fmt.Sprintf(format, args...)
		printMessage("DEBUG", msg)
	}
}

// Info prints info messages (visible at INFO level and above)
func Info(format string, args ...interface{}) {
	if shouldPrint(LevelInfo) {
		msg := fmt.Sprintf(format, args...)
		printMessage("INFO", msg)
	}
}

// Warning prints warning messages (visible at WARNING level and above)
func Warning(format string, args ...interface{}) {
	if shouldPrint(LevelWarning) {
		msg := fmt.Sprintf(format, args...)
		printMessage("WARNING", msg)
	}
}

// Error prints error messages (always visible unless log level is DISABLED)
func Error(format string, args ...interface{}) {
	if shouldPrint(LevelError) {
		msg := fmt.Sprintf(format, args...)
		printError(msg)
	}
}

// printMessage prints a formatted message with timestamp and level
func printMessage(level, message string) {
	fmt.Printf("%s\n", message)
	//fmt.Printf("[%s] %s\n", level, message)
}

// printError prints an error message in red color
func printError(message string) {
	// ANSI escape codes for red color
	const redColor = "\033[1;31m"
	const resetColor = "\033[0m"

	fmt.Printf("%s[ERROR] %s%s\n", redColor, message, resetColor)
}

// Helper functions to set specific log levels
func SetLevelDebug()   { SetLevel(LevelDebug) }
func SetLevelInfo()    { SetLevel(LevelInfo) }
func SetLevelWarning() { SetLevel(LevelWarning) }
func SetLevelError()   { SetLevel(LevelError) }
func DisableLogging()  { SetLevel(LevelDisabled) }

// SetOutput allows redirecting output (optional feature)
var output io.Writer = os.Stdout

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}
