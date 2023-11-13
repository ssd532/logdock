package main

import (
	"bytes"
	"testing"

	"github.com/ssd532/logdock/logharbour"
)

type ValidatorFunc func(entry any) error

// Ensure ValidatorFunc implements the logharbour.Validator interface by providing a Validate method.
func (vf ValidatorFunc) Validate(entry any) error {
	return vf(entry)
}

// mockWriter is a simple in-memory writer to capture log outputs for testing.
type mockWriter struct {
	bytes.Buffer
}

func (m *mockWriter) Close() error {
	return nil
}

// TestPriorityLevelPrinting checks that a more verbose priority level prints all less verbose messages.
func TestPriorityLevelPrinting(t *testing.T) {
	// Create a new mock writer to capture log outputs.
	output := new(mockWriter)

	// Create a fallback writer that uses the mock writer for both primary and fallback outputs.
	fallbackWriter := logharbour.NewFallbackWriter(output, output)

	// Initialize the logger with a basic context and validator, and a test priority level.
	logger := logharbour.NewLogger(logharbour.LoggerContext{}, ValidatorFunc(func(entry any) error {
		return nil
	}), logharbour.Debug1, fallbackWriter)

	// Log a message at Debug1 level.
	logger.LogDebug(logharbour.Debug1, "Debug1 message", logharbour.DebugInfo{})

	// Change logger priority to a more verbose level (Debug2).
	logger.ChangePriority(logharbour.Debug2)

	// Log another message at Debug2 level.
	logger.LogDebug(logharbour.Debug2, "Debug2 message", logharbour.DebugInfo{})

	// Check if both messages are present in the output.
	outputStr := output.String()
	if !bytes.Contains(output.Bytes(), []byte("Debug1 message")) {
		t.Errorf("Expected Debug1 message to be logged, got: %s", outputStr)
	}
	if !bytes.Contains(output.Bytes(), []byte("Debug2 message")) {
		t.Errorf("Expected Debug2 message to be logged, got: %s", outputStr)
	}
}
