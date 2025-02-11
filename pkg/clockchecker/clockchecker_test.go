package clockchecker

import (
	"bytes"
	"testing"
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/aukilabs/go-tooling/pkg/logs"
	"github.com/stretchr/testify/require"
)

// TestCheckClockSync verifies different log levels based on time skew.
func TestCheckClockSync(t *testing.T) {
	tests := []struct {
		name      string
		skew      time.Duration
		expectLog string
		failNTP   bool // Simulate NTP failure
	}{
		{"Clock in sync", 2 * time.Second, "system clock is in sync", false},
		{"Clock in sync (neg)", -2 * time.Second, "system clock is in sync", false},
		{"Out of sync", 6 * time.Second, "system clock is out of sync", false},
		{"Out of sync (neg)", -6 * time.Second, "system clock is out of sync", false},
		{"Severely out of sync", 12 * time.Second, "system clock is severely out of sync", false},
		{"Severely out of sync (neg)", -12 * time.Second, "system clock is severely out of sync", false},
		{"NTP failure", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logs using a buffer
			var logBuffer bytes.Buffer
			logs.SetLogger(func(entry logs.Entry) {
				logBuffer.WriteString(entry.String() + "\n")
			})

			// Create a ClockChecker instance
			clockChecker := New()

			// Override getNTPTime for test case
			if tt.failNTP {
				clockChecker.getNTPTime = func() (time.Time, error) {
					return time.Time{}, errors.New("NTP failure")
				}
			} else {
				clockChecker.getNTPTime = func() (time.Time, error) {
					return time.Now().Add(tt.skew), nil
				}
			}

			// Run clock sync check
			err := clockChecker.checkClockSync()
			logOutput := logBuffer.String()

			// Expect error only if it's an NTP failure
			if tt.failNTP {
				require.Error(t, err)
				// Check error instead of log message
				require.Contains(t, err.Error(), "NTP failure")
			} else {
				require.NoError(t, err)
			}

			// Verify expected log message
			require.Contains(t, logOutput, tt.expectLog)
		})
	}
}
