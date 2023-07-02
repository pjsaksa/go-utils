package log

import (
	"strings"
	"testing"
	"time"
)

func TestSimpleLogs(t *testing.T) {
	// Setup
	var out strings.Builder
	SetOutput(&out)
	setForcedTime(time.Unix(42000000, 42000000))

	// Cleanup
	defer func() {
		ResetOutput()
		resetForcedTime()
	}()

	// Test object
	const logMessage = "test log message"
	DEBUG(logMessage)
	ERROR(logMessage)
	EVENT(logMessage)
	INFO(logMessage)
	WARNING(logMessage)

	// Verify output.
	const expected = "04:40:00.042 " + debugColor + logMessage + resetColor + "\n" +
		"04:40:00.042 " + errorColor + "ERROR: " + logMessage + resetColor + "\n" +
		"04:40:00.042 " + eventColor + "EVENT: " + logMessage + resetColor + "\n" +
		"04:40:00.042 " + infoColor + "INFO: " + logMessage + resetColor + "\n" +
		"04:40:00.042 " + warningColor + "WARNING: " + logMessage + resetColor + "\n"
	if out.String() != expected {
		t.Errorf("Output does not match expected:\nWANT:\n%s\nGOT:\n%s",
			expected,
			out.String())
	}
}
