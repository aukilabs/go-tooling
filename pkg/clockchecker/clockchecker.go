package clockchecker

import (
	"context"
	"time"

	"github.com/aukilabs/go-tooling/pkg/errors"
	"github.com/aukilabs/go-tooling/pkg/logs"
	"github.com/beevik/ntp"
)

const (
	DefaultInitialDelay     = 3 * time.Second
	DefaultSecondCheckDelay = 1 * time.Minute
	DefaultCheckInterval    = 6 * time.Hour
	DefaultWarningThreshold = 5 * time.Second
	DefaultErrorThreshold   = 10 * time.Second
	DefaultNTPServerAddress = "pool.ntp.org"
)

type ClockChecker interface {
	// Start runs the clock synchronization check at configured intervals in a goroutine.
	Start(ctx context.Context)
}

type Options struct {
	InitialDelay     time.Duration
	SecondCheckDelay time.Duration
	CheckInterval    time.Duration
	WarningThreshold time.Duration
	ErrorThreshold   time.Duration
	NTPServerAddress string
}

type clockChecker struct {
	started          bool
	initialDelay     time.Duration
	secondCheckDelay time.Duration
	checkInterval    time.Duration
	warningThreshold time.Duration
	errorThreshold   time.Duration
	getNTPTime       func() (time.Time, error)
}

func New(opts *Options) *clockChecker {
	cc := &clockChecker{
		initialDelay:     DefaultInitialDelay,
		secondCheckDelay: DefaultSecondCheckDelay,
		checkInterval:    DefaultCheckInterval,
		warningThreshold: DefaultWarningThreshold,
		errorThreshold:   DefaultErrorThreshold,
		getNTPTime: func() (time.Time, error) {
			return ntp.Time(DefaultNTPServerAddress)
		},
	}

	if opts == nil {
		return cc
	}
	if opts.InitialDelay != time.Duration(0) {
		cc.initialDelay = opts.InitialDelay
	}
	if opts.SecondCheckDelay != time.Duration(0) {
		cc.secondCheckDelay = opts.SecondCheckDelay
	}
	if opts.CheckInterval != time.Duration(0) {
		cc.checkInterval = opts.CheckInterval
	}
	if opts.WarningThreshold != time.Duration(0) {
		cc.warningThreshold = opts.WarningThreshold
	}
	if opts.ErrorThreshold != time.Duration(0) {
		cc.errorThreshold = opts.ErrorThreshold
	}
	if opts.NTPServerAddress != "" {
		cc.getNTPTime = func() (time.Time, error) {
			return ntp.Time(opts.NTPServerAddress)
		}
	}

	return cc
}

// checkClockSync checks if the system clock is out of sync and logs at different levels.
func (c *clockChecker) checkClockSync() error {
	ntpTime, err := c.getNTPTime()
	if err != nil {
		return errors.New("failed to retrieve time from NTP server").Wrap(err)
	}

	localTime := time.Now()
	diff := localTime.Sub(ntpTime)

	// Log based on thresholds.
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

// Start runs the clock synchronization check at configured intervals in a goroutine.
func (c *clockChecker) Start(ctx context.Context) {
	if c.started {
		return
	}
	c.started = true

	go func() {
		// Initial check (default after 3 seconds).
		time.Sleep(c.initialDelay)
		if err := c.checkClockSync(); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Second check (default after 1 minute).
		time.Sleep(c.secondCheckDelay - c.initialDelay)
		if err := c.checkClockSync(); err != nil {
			logs.Error(errors.New("clock skew check failed").Wrap(err))
		}

		// Periodic checks (default every 6 hours).
		ticker := time.NewTicker(c.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.checkClockSync(); err != nil {
					logs.Error(errors.New("clock skew check failed").Wrap(err))
				}
			}
		}
	}()
}
