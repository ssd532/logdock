package logharbour

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const defaultPriority = Info

type Logger struct {
	appName        string
	system         string // Add the system field
	priority       logPriority
	who            string
	remoteIP       string
	module         string // Add the module field
	op             string // Add the operation field
	whatClass      string // Add the whatClass field
	whatInstanceId string // Add the whatInstanceId field
	status         Status
	writer         io.Writer
	validator      Validator
	mu             sync.Mutex
}

func (l *Logger) clone() *Logger {
	return &Logger{
		appName:        l.appName,
		system:         l.system,
		writer:         l.writer,
		validator:      l.validator,
		priority:       l.priority,
		who:            l.who,
		remoteIP:       l.remoteIP,
		module:         l.module,
		op:             l.op,
		whatClass:      l.whatClass,
		whatInstanceId: l.whatInstanceId,
		status:         l.status,
	}
}

func NewLogger(appName string, validator Validator, writer io.Writer) *Logger {
	return &Logger{
		appName:   appName,
		system:    GetSystemName(),
		writer:    writer,
		validator: validator,
		priority:  defaultPriority,
	}
}

func (l *Logger) WithWho(who string) *Logger {
	newLogger := l.clone() // Create a copy of the logger
	newLogger.who = who    // Change the 'who' field
	return newLogger       // Return the new logger
}

func (l *Logger) WithModule(module string) *Logger {
	newLogger := l.clone()
	newLogger.module = module
	return newLogger
}

func (l *Logger) WithOp(op string) *Logger {
	newLogger := l.clone()
	newLogger.op = op
	return newLogger
}

func (l *Logger) WithWhatClass(whatClass string) *Logger {
	newLogger := l.clone()
	newLogger.whatClass = whatClass
	return newLogger
}

func (l *Logger) WithWhatInstanceId(whatInstanceId string) *Logger {
	newLogger := l.clone()
	newLogger.whatInstanceId = whatInstanceId
	return newLogger
}

func (l *Logger) WithStatus(status Status) *Logger {
	newLogger := l.clone()
	newLogger.status = status
	return newLogger
}

func (l *Logger) WithPriority(priority logPriority) *Logger {
	newLogger := l.clone()
	newLogger.priority = priority
	return newLogger
}

func (l *Logger) WithRemoteIP(remoteIP string) *Logger {
	newLogger := l.clone()
	newLogger.remoteIP = remoteIP
	return newLogger
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

func (l *Logger) newLogEntry(message string, data interface{}) LogEntry {
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

func (l *Logger) LogDataChange(message string, data ChangeInfo) error {
	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeChange
	return l.log(entry)
}

func (l *Logger) LogActivity(message string, data ActivityInfo) error {
	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeActivity
	return l.log(entry)
}

func (l *Logger) LogDebug(message string, data DebugInfo) error {
	data.FileName, data.LineNumber, data.FunctionName, data.StackTrace = GetDebugInfo(2)
	data.Pid = os.Getpid()
	data.Runtime = runtime.Version()

	entry := l.newLogEntry(message, data)
	entry.Type = LogTypeDebug
	return l.log(entry)
}

func (l *Logger) ChangePriority(newPriority logPriority) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.priority = newPriority
}

func (lp logPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(lp.string())
}

func (lt LogType) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.string())
}
