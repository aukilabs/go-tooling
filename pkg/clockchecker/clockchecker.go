package clockchecker

import (
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/aukilabs/go-tooling/pkg/logs"
	"github.com/beevik/ntp"
)

// Configurable parameters
var (
	ntpServer        = "pool.ntp.org"
	initialDelay     = 3 * time.Second // First check after 3 seconds
	secondCheckDelay = 1 * time.Minute // Second check after 1 minute
	checkInterval    = 6 * time.Hour   // Subsequent checks every 6 hours

	warningThreshold = 5 * time.Second  // Warning if out of sync by more than 5s
	errorThreshold   = 10 * time.Second // Error if out of sync by more than 10s

	// Function variable for testing
	getNTPTime = func() (time.Time, error) {
		return ntp.Time(ntpServer)
	}
)

// CheckClockSync checks if the system clock is out of sync and logs at different levels.
func CheckClockSync() error {
	ntpTime, err := getNTPTime()
	if err != nil {
		logs.Error(errors.New("failed to retrieve time from NTP server").Wrap(err))
		return err
	}

	localTime := time.Now()
	diff := localTime.Sub(ntpTime)

	// Log based on thresholds
	switch {
	case diff > errorThreshold || diff < -errorThreshold:
		logs.WithTag("difference", diff).
			WithTag("local_time", localTime).
			WithTag("ntp_time", ntpTime).
			Error(errors.New("system clock is severely out of sync"))
	case diff > warningThreshold || diff < -warningThreshold:
		logs.WithTag("difference", diff).
			WithTag("local_time", localTime).
			WithTag("ntp_time", ntpTime).
			Warn("system clock is out of sync")
	default:
		logs.WithTag("difference", diff).
			Info("system clock is in sync")
	}

	return nil
}

// StartSyncMonitor runs the clock synchronization check at configured intervals in a goroutine.
func StartSyncMonitor() {
	go func() {
		// Initial check after 3 seconds
		time.Sleep(initialDelay)
		if err := CheckClockSync(); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Second check after 1 minute
		time.Sleep(secondCheckDelay - initialDelay)
		if err := CheckClockSync(); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Periodic checks every 6 hours
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := CheckClockSync(); err != nil {
				logs.Error(errors.New("clock skew check failed").Wrap(err))
			}
		}
	}()
}
