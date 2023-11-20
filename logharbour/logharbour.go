package logharbour

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

const defaultPriority = Info

// Logger provides a structured interface for logging.
// It's designed for each goroutine to have its own instance.
// Logger is safe for concurrent use. However, it's not recommended
// to share a Logger instance across multiple goroutines.
type Logger struct {
	appName        string              // Name of the application.
	system         string              // System where the application is running.
	priority       logPriority         // Priority level of the log messages.
	who            string              // User or service performing the operation.
	remoteIP       string              // IP address of the remote endpoint.
	module         string              // Module or subsystem within the application.
	op             string              // Operation being performed.
	whatClass      string              // Class of the object instance involved.
	whatInstanceId string              // Unique ID of the object instance.
	status         Status              // Status of the operation.
	writer         io.Writer           // Writer interface for log entries.
	validator      *validator.Validate // Validator for log entries.
	mu             sync.Mutex          // Mutex for thread-safe operations.
}

// clone creates and returns a new Logger with the same values as the original.
func (l *Logger) clone() *Logger {
	return &Logger{
		appName:        l.appName,
		system:         l.system,
		writer:         l.writer,
		priority:       l.priority,
		who:            l.who,
		remoteIP:       l.remoteIP,
		module:         l.module,
		op:             l.op,
		whatClass:      l.whatClass,
		whatInstanceId: l.whatInstanceId,
		status:         l.status,
		validator:      l.validator,
	}
}

// NewLogger creates a new Logger with the specified application name and writer.
func NewLogger(appName string, writer io.Writer) *Logger {
	return &Logger{
		appName:   appName,
		system:    GetSystemName(),
		writer:    writer,
		validator: validator.New(),
		priority:  defaultPriority,
	}
}

// WithWho returns a new Logger with the 'who' field set to the specified value.
func (l *Logger) WithWho(who string) *Logger {
	newLogger := l.clone() // Create a copy of the logger
	newLogger.who = who    // Change the 'who' field
	return newLogger       // Return the new logger
}

// WithModule returns a new Logger with the 'module' field set to the specified value.
func (l *Logger) WithModule(module string) *Logger {
	newLogger := l.clone()
	newLogger.module = module
	return newLogger
}

// WithOp returns a new Logger with the 'op' field set to the specified value.
func (l *Logger) WithOp(op string) *Logger {
	newLogger := l.clone()
	newLogger.op = op
	return newLogger
}

// WithWhatClass returns a new Logger with the 'whatClass' field set to the specified value.
func (l *Logger) WithWhatClass(whatClass string) *Logger {
	newLogger := l.clone()
	newLogger.whatClass = whatClass
	return newLogger
}

// WithWhatInstanceId returns a new Logger with the 'whatInstanceId' field set to the specified value.
func (l *Logger) WithWhatInstanceId(whatInstanceId string) *Logger {
	newLogger := l.clone()
	newLogger.whatInstanceId = whatInstanceId
	return newLogger
}

// WithStatus returns a new Logger with the 'status' field set to the specified value.
func (l *Logger) WithStatus(status Status) *Logger {
	newLogger := l.clone()
	newLogger.status = status
	return newLogger
}

// WithPriority returns a new Logger with the 'priority' field set to the specified value.
func (l *Logger) WithPriority(priority logPriority) *Logger {
	newLogger := l.clone()
	newLogger.priority = priority
	return newLogger
}

// WithRemoteIP returns a new Logger with the 'remoteIP' field set to the specified value.
func (l *Logger) WithRemoteIP(remoteIP string) *Logger {
	newLogger := l.clone()
	newLogger.remoteIP = remoteIP
	return newLogger
}

// log writes a log entry. It locks the Logger's mutex to prevent concurrent write operations.
func (l *Logger) log(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.AppName = l.appName
	if !l.shouldLog(entry.Priority) {
		return nil
	}
	if err := l.validator.Struct(entry); err != nil {
		return err
	}
	return formatAndWriteEntry(l.writer, entry)
}

// shouldLog determines whether a log entry should be written based on its priority.
func (l *Logger) shouldLog(p logPriority) bool {
	return p >= l.priority
}

// formatAndWriteEntry formats a log entry as JSON and writes it to the Logger's writer.
func formatAndWriteEntry(writer io.Writer, entry LogEntry) error {
	formattedEntry, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	formattedEntry = append(formattedEntry, '\n')
	_, writeErr := writer.Write(formattedEntry)
	return writeErr
}

// newLogEntry creates a new log entry with the specified message and data.
func (l *Logger) newLogEntry(message string, data any) LogEntry {
	return LogEntry{
		AppName:        l.appName,
		System:         l.system,
		Priority:       l.priority,
		When:           time.Now().UTC(),
		Message:        message,
		Data:           data,
		Who:            l.who,
		RemoteIP:       l.remoteIP,
		Module:         l.module,         // Add the module field
		Op:             l.op,             // Add the operation field
		WhatClass:      l.whatClass,      // Add the whatClass field
		WhatInstanceId: l.whatInstanceId, // Add the whatInstanceId field
		Status:         l.status,
	}
}

// LogDataChange logs a data change event.
func (l *Logger) LogDataChange(message string, data ChangeInfo) error {
	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeChange
	return l.log(entry)
}

// LogActivity logs an activity event.
func (l *Logger) LogActivity(message string, data ActivityInfo) error {
	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeActivity
	return l.log(entry)
}

// LogDebug logs a debug event.
func (l *Logger) LogDebug(message string, data DebugInfo) error {
	data.FileName, data.LineNumber, data.FunctionName, data.StackTrace = GetDebugInfo(2)
	data.Pid = os.Getpid()
	data.Runtime = runtime.Version()

	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeDebug
	return l.log(entry)
}

// ChangePriority changes the priority level of the Logger.
func (l *Logger) ChangePriority(newPriority logPriority) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.priority = newPriority
}
