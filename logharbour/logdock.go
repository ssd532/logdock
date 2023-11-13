package logharbour

import (
	"encoding/json"
	"io"
	"time"
)

// LogPriority defines the severity level of a log message.
type LogPriority int

const (
	// Debug2 represents extremely verbose debugging information.
	Debug2 LogPriority = iota + 1
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

// String returns the string representation of the LogPriority.
func (lp LogPriority) String() string {
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

// MarshalJSON customizes the JSON representation of the LogPriority.
func (lp LogPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(lp.String())
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
func (lt LogType) String() string {
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

// MarshalJSON customizes the JSON representation of the LogType.
func (lt LogType) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.String())
}

// LogEntry encapsulates all the relevant information for a log message.
type LogEntry struct {
	Context   LoggerContext // The fixed set of fields that should be included in every log entry.
	Type      LogType       // The category of the log entry.
	Priority  LogPriority   // The severity level of the log entry.
	Timestamp time.Time     // The time at which the log entry was created.
	Message   string        // A descriptive message for the log entry.
	Data      any           // The payload of the log entry, can be any type.
}

// Validator defines the interface for log entry validation.
type Validator interface {
	Validate(any) error
}

// Logger defines the interface for logging operations.
type Logger interface {
	Log(entry LogEntry) error
}

// HarbourLogger is the main logger with configurable priority and validation.
type HarbourLogger struct {
	context   LoggerContext
	writer    io.Writer
	validator Validator
	priority  LogPriority
}

// LoggerContext holds the fixed set of fields that should be included in every log entry.
type LoggerContext struct {
	AppName    string
	SystemName string
}

// NewHarbourLogger initializes a new HarbourLogger with the provided configuration.
func NewLogger(context LoggerContext, validator Validator, priority LogPriority, writer io.Writer) *HarbourLogger {
	return &HarbourLogger{
		context:   context,
		writer:    writer,
		validator: validator,
		priority:  priority,
	}
}

// Log processes and logs the given LogEntry if it meets the priority requirements.
func (l *HarbourLogger) Log(entry LogEntry) error {
	if !l.ShouldLog(entry.Priority) {
		return nil // Do not log if the entry's priority is below the logger's threshold.
	}
	if err := l.validator.Validate(entry); err != nil {
		return err // Return validation error.
	}

	// Format the log entry and write it using the logger's writer.
	// The formatting and writing logic will be implemented in the next steps.
	// For now, we'll assume a function `formatAndWriteEntry` exists to handle this.
	return formatAndWriteEntry(l.writer, entry)
}

// ShouldLog determines if a log entry should be logged based on its priority.
func (l *HarbourLogger) ShouldLog(p LogPriority) bool {
	return p >= l.priority
}

// formatAndWriteEntry formats the log entry as JSON and writes it to the provided writer.
func formatAndWriteEntry(writer io.Writer, entry LogEntry) error {
	formattedEntry, err := json.Marshal(entry)
	if err != nil {
		return err // Return error if marshaling fails.
	}
	formattedEntry = append(formattedEntry, '\n') // Add newline for readability.
	_, writeErr := writer.Write(formattedEntry)
	return writeErr // Return error if writing fails.
}

// ChangeInfo holds information about data changes such as creations, updates, or deletions.
type ChangeInfo struct {
	Entity    string                 `json:"entity"`
	Operation string                 `json:"operation"`
	User      string                 `json:"user"`
	Changes   map[string]interface{} `json:"changes"`
}

// ActivityInfo holds information about system activities like web service calls or function executions.
type ActivityInfo struct {
	ActivityType string        `json:"activityType"`
	Endpoint     string        `json:"endpoint"`
	Duration     time.Duration `json:"duration"`
}

// DebugInfo holds debugging information that can help in software diagnostics.
type DebugInfo struct {
	Level    string `json:"level"`
	Message  string `json:"message"`
	Location string `json:"location"` // Could be a file name, line number, etc.
}

// LogDataChange logs an entry related to data changes with the specified priority and message.
func (l *HarbourLogger) LogDataChange(priority LogPriority, message string, data ChangeInfo) error {
	return l.Log(LogEntry{
		Context:   l.context,
		Type:      LogTypeChange,
		Priority:  priority,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	})
}

// LogActivity logs an entry related to activities with the specified priority and message.
func (l *HarbourLogger) LogActivity(priority LogPriority, message string, data ActivityInfo) error {
	return l.Log(LogEntry{
		Context:   l.context,
		Type:      LogTypeActivity,
		Priority:  priority,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	})
}

// LogDebug logs an entry related to debugging with the specified priority and message.
func (l *HarbourLogger) LogDebug(priority LogPriority, message string, data DebugInfo) error {
	return l.Log(LogEntry{
		Context:   l.context,
		Type:      LogTypeDebug,
		Priority:  priority,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	})
}

// FallbackWriter provides an io.Writer that automatically falls back to a secondary writer if the primary writer fails.
type FallbackWriter struct {
	primary  io.Writer // The main writer to which log entries will be written.
	fallback io.Writer // The fallback writer used if the primary writer fails.
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
	n, err = fw.primary.Write(p)
	if err != nil {
		// Primary writer failed; attempt to write to the fallback writer.
		n, err = fw.fallback.Write(p)
	}
	return n, err // Return the result of the write operation.
}

func (l *HarbourLogger) ChangePriority(newPriority LogPriority) {
	l.priority = newPriority
}
