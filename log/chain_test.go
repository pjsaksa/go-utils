package log

import (
	"strings"
	"testing"
	"time"
)

func TestChain_1(t *testing.T) {
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
	const logMessage1 = "HTTP Request"
	const logMessage2 = "404"
	chain1 := Chain(EventMsg(logMessage1))
	chain1.Add(WarningMsg(logMessage2))
	chain1.Write()

	// Verify output.
	const expected = "04:40:00.042 " + eventColor + "EVENT: " + logMessage1 + resetColor + " -> " +
		warningColor + "WARNING: " + logMessage2 + resetColor + "\n"
	if out.String() != expected {
		t.Errorf("Output does not match expected:\nWANT:\n%s\nGOT:\n%s",
			expected,
			out.String())
	}
}
