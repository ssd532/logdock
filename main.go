package main

import (
	"os"
	"time"

	"github.com/ssd532/logdock/logharbour"
)

type ValidatorFunc func(entry any) error

func (vf ValidatorFunc) Validate(entry any) error {
	return vf(entry.(logharbour.LogEntry))
}

func main() {
	// Define the common logger context.
	context := logharbour.LoggerContext{
		AppName:    "MyAwesomeApp",
		SystemName: "MyAwesomeSystem",
	}

	// Define a simple validator that always returns nil for no error.
	validator := ValidatorFunc(func(entry any) error {
		return nil
	})

	// Create a fallback writer that uses stdout as the fallback.
	fallbackWriter := logharbour.NewFallbackWriter(os.Stdout, os.Stdout)

	// Initialize the logger with the context, validator, default priority, and fallback writer.
	logger := logharbour.NewLogger(context, validator, logharbour.Info, fallbackWriter)

	// Log an activity entry.
	logger.LogActivity(logharbour.Info, "User logged in", logharbour.ActivityInfo{
		ActivityType: "UserLogin",
		Endpoint:     "/api/v1/login",
		Duration:     120 * time.Millisecond,
	})

	// Log a data change entry.
	logger.LogDataChange(logharbour.Info, "User updated profile", logharbour.ChangeInfo{
		Entity:    "User",
		Operation: "Update",
		User:      "johndoe",
		Changes:   map[string]interface{}{"email": "john@example.com"},
	})

	// Log a debug entry.
	logger.LogDebug(logharbour.Debug1, "Debugging user session", logharbour.DebugInfo{
		Level:    "DEBUG1",
		Message:  "Session ID is valid",
		Location: "session_manager.go:45",
	})

	// Change logger priority at runtime.
	logger.ChangePriority(logharbour.Debug2)

	// Log another debug entry with a higher verbosity level.
	logger.LogDebug(logharbour.Debug2, "Detailed debugging info", logharbour.DebugInfo{
		Level:    "DEBUG2",
		Message:  "Trace: start session renewal process",
		Location: "session_manager.go:50",
	})
}
