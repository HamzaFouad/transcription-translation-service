package utils

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

type BackoffConfig struct {
	InitialInterval time.Duration
	Multiplier      float64
	MaxElapsedTime  time.Duration
	MaxInterval     time.Duration
	Notify          func(err error, duration time.Duration)
}

func NewBackoffConfig(logger Logger) *BackoffConfig {

	defaultNotifier := func(err error, duration time.Duration) {
		logger.Warn("Retry after %s due to error: %v", duration, err)
	}

	return &BackoffConfig{
		InitialInterval: 500 * time.Millisecond, // default initial interval
		Multiplier:      2.0,                    // default multiplier
		MaxElapsedTime:  30 * time.Second,       // default max elapsed time
		MaxInterval:     5 * time.Second,        // default max interval
		Notify:          defaultNotifier,        // set the default notifier function

	}
}

func (cfg *BackoffConfig) ToBackOff() *backoff.ExponentialBackOff {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = cfg.InitialInterval
	expBackoff.Multiplier = cfg.Multiplier
	expBackoff.MaxElapsedTime = cfg.MaxElapsedTime
	expBackoff.MaxInterval = cfg.MaxInterval
	return expBackoff
}
