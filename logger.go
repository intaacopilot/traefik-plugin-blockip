package traefik_plugin_blockip

import (
	"fmt"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger handles logging for the plugin
type Logger struct {
	debug       bool
	mu          sync. Mutex
	logBuffer   []string
	maxBuffSize int
}

// NewLogger creates a new logger instance
func NewLogger(debug bool) *Logger {
	return &Logger{
		debug:       debug,
		logBuffer:   make([]string, 0),
		maxBuffSize: 1000,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debug {
		l. log("DEBUG", format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// log is the internal logging method
func (l *Logger) log(level string, format string, args ...interface{}) {
	l.mu. Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf("[%s] %s - %s", time.Now().Format("2006-01-02 15:04:05"), level, fmt.Sprintf(format, args... ))
	
	// Print to stdout/stderr
	fmt.Println(message)

	// Store in buffer
	if len(l.logBuffer) < l.maxBuffSize {
		l.logBuffer = append(l.logBuffer, message)
	} else {
		// Rotate buffer
		l.logBuffer = append(l.logBuffer[1:], message)
	}
}

// GetLogs retrieves recent logs
func (l *Logger) GetLogs(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if count <= 0 || count > len(l.logBuffer) {
		return l.logBuffer
	}

	return l. logBuffer[len(l.logBuffer)-count:]
}

// ClearLogs clears the log buffer
func (l *Logger) ClearLogs() {
	l.mu. Lock()
	defer l.mu.Unlock()

	l. logBuffer = make([]string, 0)
}