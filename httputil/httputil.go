package httputil

import (
	"math"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
)

// DoWithRetries decorates a function that returns (T, Error) to retry with exponential backoff.
// Zero-value retryOn and backoffConfigs will be replaced with sane defaults.
func DoWithRetries[T any](
	call func() (T, app.Error),
	maxRetries int,
	retryOn func(statusCode int) bool,
	backoffConfigs ExponentialBackoffConfigs,
) (v T, err app.Error) {
	// set defaults
	if retryOn == nil {
		retryOn = func(code int) bool {
			return code == 502
		}
	}
	if backoffConfigs.Interval == 0 {
		backoffConfigs.Interval = time.Second
	}
	if backoffConfigs.Max == 0 {
		backoffConfigs.Max = 10 * time.Minute
	}
	if backoffConfigs.Rate == 0 {
		backoffConfigs.Rate = 3
	}

	//
	for i := 0; i < maxRetries; i++ {
		v, err = call()
		if err != nil {
			if retryOn(err.StatusCode()) {
				time.Sleep(exponentialBackoff(backoffConfigs, i))
			}
			continue
		}
		break
	}
	return v, err
}

type ExponentialBackoffConfigs struct {
	Rate                   float64
	Interval, Initial, Max time.Duration
}

func exponentialBackoff(
	cfg ExponentialBackoffConfigs,
	iteration int,
) time.Duration {
	backoff := time.Duration(math.Pow(cfg.Rate, float64(iteration)))*cfg.Interval + cfg.Initial
	return min(backoff, cfg.Max)
}
