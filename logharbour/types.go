package logharbour

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// logPriority defines the severity level of a log message.
type logPriority int

const (
	// Debug2 represents extremely verbose debugging information.
	Debug2 logPriority = iota + 1
	// Debug1 represents detailed debugging information.
	Debug1
	// Debug0 represents high-level debugging information.
	Debug0
	// Info represents informational messages.
	Info
	// Warn represents warning messages.
	Warn
	// Err represents error messages where operations failed to complete.
	Err
	// Crit represents critical failure messages.
	Crit
	// Sec represents security alert messages.
	Sec
)

// String returns the string representation of the logPriority.
func (lp logPriority) string() string {
	switch lp {
	case Debug2:
		return "Debug2"
	case Debug1:
		return "Debug1"
	case Debug0:
		return "Debug0"
	case Info:
		return "Info"
	case Warn:
		return "Warn"
	case Err:
		return "Err"
	case Crit:
		return "Crit"
	case Sec:
		return "Sec"
	default:
		return "Unknown"
	}
}

// MarshalJSON is required by the encoding/json package.
// It converts the logPriority to its string representation and returns it as a JSON-encoded value.
func (lp logPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(lp.string())
}

// LogType defines the category of a log message.
type LogType int

const (
	// LogTypeChange represents a log entry for data changes.
	LogTypeChange LogType = iota + 1
	// LogTypeActivity represents a log entry for activities such as web service calls.
	LogTypeActivity
	// LogTypeDebug represents a log entry for debug information.
	LogTypeDebug
)

// String returns the string representation of the LogType.
func (lt LogType) string() string {
	switch lt {
	case LogTypeChange:
		return "Change"
	case LogTypeActivity:
		return "Activity"
	case LogTypeDebug:
		return "Debug"
	default:
		return "Unknown"
	}
}

// MarshalJSON is required by the encoding/json package.
// It converts the LogType to its string representation and returns it as a JSON-encoded value.
func (lt LogType) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.string())
}

type Status int

const (
	Success Status = iota
	Failure
)

// LogEntry encapsulates all the relevant information for a log message.
type LogEntry struct {
	AppName        string // The name of the application.
	System         string
	Type           LogType     // The category of the log entry.
	Priority       logPriority // The severity level of the log entry.
	Who            string
	When           time.Time // The time at which the log entry was created.
	Op             string    // The operation being performed
	WhatClass      string    // The unique ID, name of the object instance on which the operation was being attempted
	WhatInstanceId string    // The unique ID, name, or other "primary key" information of the object instance on which the operation was being attempted
	Status         Status    // 0 or 1, indicating success (1) or failure (0), or some other binary representation
	RemoteIP       string    // IP address of remote endpoint from where the operation is being performed.
	Message        string    // A descriptive message for the log entry.
	Data           any       // The payload of the log entry, can be any type.
	Module         string    // The module or subsystem within the application
}

// ChangeInfo holds information about data changes such as creations, updates, or deletions.
type ChangeInfo struct {
	Entity    string         `json:"entity"`
	Operation string         `json:"operation"`
	Changes   map[string]any `json:"changes"`
}

// ActivityInfo holds information about system activities like web service calls or function executions.
type ActivityInfo any

// DebugInfo holds debugging information that can help in software diagnostics.
type DebugInfo struct {
	Pid          int            `json:"pid"`
	Runtime      string         `json:"runtime"`
	FileName     string         `json:"fileName"`
	LineNumber   int            `json:"lineNumber"`
	FunctionName string         `json:"functionName"`
	StackTrace   string         `json:"stackTrace"`
	Variables    map[string]any `json:"variables"`
}

// FallbackWriter provides an io.Writer that automatically falls back to a secondary writer if the primary writer fails.
type FallbackWriter struct {
	primary  io.Writer // The main writer to which log entries will be written.
	fallback io.Writer // The fallback writer used if the primary writer fails.
	mu       sync.Mutex
}

// NewFallbackWriter creates a new FallbackWriter with a specified primary and fallback writer.
func NewFallbackWriter(primary, fallback io.Writer) *FallbackWriter {
	return &FallbackWriter{
		primary:  primary,
		fallback: fallback,
	}
}

// Write attempts to write the byte slice to the primary writer, falling back to the secondary writer on error.
// It returns the number of bytes written and any error encountered that caused the write to stop early.
func (fw *FallbackWriter) Write(p []byte) (n int, err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	n, err = fw.primary.Write(p)
	if err != nil {
		// Primary writer failed; attempt to write to the fallback writer.
		n, err = fw.fallback.Write(p)
	}
	return n, err // Return the result of the write operation.
}
