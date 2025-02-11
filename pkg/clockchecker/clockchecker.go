package clockchecker

import (
	"context"
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/aukilabs/go-tooling/pkg/logs"
	"github.com/beevik/ntp"
)

type ClockChecker struct {
	initialDelay     time.Duration
	secondCheckDelay time.Duration
	checkInterval    time.Duration
	warningThreshold time.Duration
	errorThreshold   time.Duration
	getNTPTime       func() (time.Time, error)
}

func New() *ClockChecker {
	return &ClockChecker{
		initialDelay:     3 * time.Second,
		secondCheckDelay: 1 * time.Minute,
		checkInterval:    6 * time.Hour,
		warningThreshold: 5 * time.Second,
		errorThreshold:   10 * time.Second,
		getNTPTime: func() (time.Time, error) {
			return ntp.Time("pool.ntp.org")
		},
	}
}

// checkClockSync checks if the system clock is out of sync and logs at different levels.
func (c *ClockChecker) checkClockSync(ctx context.Context) error {
	ntpTime, err := c.getNTPTime()
	if err != nil {
		return errors.New("failed to retrieve time from NTP server").Wrap(err)
	}

	localTime := time.Now()
	diff := localTime.Sub(ntpTime)

	// Log based on thresholds
	switch {
	case diff > c.errorThreshold || diff < -c.errorThreshold:
		logs.WithTag("difference", diff).
			WithTag("local_time", localTime).
			WithTag("ntp_time", ntpTime).
			Error(errors.New("system clock is severely out of sync"))
	case diff > c.warningThreshold || diff < -c.warningThreshold:
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
func StartSyncMonitor(ctx context.Context) {
	clockChecker := New()
	go func() {
		// Initial check after 3 seconds
		time.Sleep(clockChecker.initialDelay)
		if err := clockChecker.checkClockSync(ctx); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Second check after 1 minute
		time.Sleep(clockChecker.secondCheckDelay - clockChecker.initialDelay)
		if err := clockChecker.checkClockSync(ctx); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Periodic checks every 6 hours
		ticker := time.NewTicker(clockChecker.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := clockChecker.checkClockSync(ctx); err != nil {
					logs.Error(errors.New("clock skew check failed").Wrap(err))
				}
			}
		}
	}()
}
