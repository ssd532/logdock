package logharbour

import (
	"encoding/json"
	"io"
	"runtime"
	"strings"
	"time"
)

const defaultPriority = Info

type Logger struct {
	appName   string
	writer    io.Writer
	validator Validator
	priority  logPriority
	who       string
}

func NewLogger(appName string, validator Validator, writer io.Writer) *Logger {
	return &Logger{
		appName:   appName,
		writer:    writer,
		validator: validator,
		priority:  defaultPriority,
	}
}

func (l *Logger) WithWho(who string) *Logger {
	newLogger := *l     // Create a copy of the logger
	newLogger.who = who // Change the 'who' field
	return &newLogger   // Return the new logger
}

func (l *Logger) WithPriority(priority logPriority) *Logger {
	newLogger := *l
	newLogger.priority = priority
	return &newLogger
}

func (l *Logger) log(entry LogEntry) error {
	entry.AppName = l.appName
	if !l.shouldLog(entry.Priority) {
		return nil
	}
	if err := l.validator.Validate(entry); err != nil {
		return err
	}
	return formatAndWriteEntry(l.writer, entry)
}

func (l *Logger) shouldLog(p logPriority) bool {
	return p >= l.priority
}

func formatAndWriteEntry(writer io.Writer, entry LogEntry) error {
	formattedEntry, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	formattedEntry = append(formattedEntry, '\n')
	_, writeErr := writer.Write(formattedEntry)
	return writeErr
}

func (l *Logger) newLogEntry(priority logPriority, message string, data interface{}) LogEntry {
	return LogEntry{
		AppName:  l.appName,
		Priority: priority,
		When:     time.Now().UTC(),
		Message:  message,
		Data:     data,
		Who:      l.who,
	}
}

func (l *Logger) LogDataChange(priority logPriority, message string, data ChangeInfo) error {
	entry := l.newLogEntry(priority, message, data)
	entry.Type = LogTypeChange
	return l.log(entry)
}

func (l *Logger) LogActivity(priority logPriority, message string, data ActivityInfo) error {
	entry := l.newLogEntry(priority, message, data)
	entry.Type = LogTypeActivity
	return l.log(entry)
}

func (l *Logger) LogDebug(priority logPriority, message string, data DebugInfo) error {
	// Get the caller info (skip 2 levels to skip the LogDebug and log functions)
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		data.FileName = file
		data.LineNumber = line

		// Get the function name
		funcName := runtime.FuncForPC(pc).Name()
		// Trim the package name
		funcName = funcName[strings.LastIndex(funcName, ".")+1:]
		data.FunctionName = funcName

		// Get the stack trace
		buf := make([]byte, 1024)
		length := runtime.Stack(buf, false)
		data.StackTrace = string(buf[:length])
	}

	entry := l.newLogEntry(priority, message, data)
	entry.Type = LogTypeDebug
	return l.log(entry)
}

func (l *Logger) ChangePriority(newPriority logPriority) {
	l.priority = newPriority
}

func (lp logPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(lp.string())
}

func (lt LogType) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.string())
}
