package grok

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/GuDaStudio/GrokSearch/internal/config"
)

func withRetry(ctx context.Context, cfg config.Config, fn func() (*http.Response, error)) error {
	attempts := cfg.RetryMaxAttempts + 1
	if attempts < 1 {
		attempts = 1
	}

	var last error
	for i := 0; i < attempts; i++ {
		resp, err := fn()
		if err == nil {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			return nil
		}
		last = err
		if i == attempts-1 || !IsRetryable(err) {
			return err
		}
		wait := retryDelay(cfg, err, i)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return last
}

func retryDelay(cfg config.Config, err error, attempt int) time.Duration {
	var status *HTTPStatusError
	if errors.As(err, &status) && status.StatusCode == 429 {
		if d, ok := parseRetryAfter(status.RetryAfter); ok {
			return d
		}
	}

	multiplier := cfg.RetryMultiplier
	if multiplier <= 0 {
		multiplier = 1
	}
	seconds := multiplier * math.Pow(2, float64(attempt))
	delay := time.Duration(seconds * float64(time.Second))
	maxWait := cfg.RetryMaxWaitDuration()
	if delay > maxWait {
		return maxWait
	}
	if delay < 100*time.Millisecond {
		return 100 * time.Millisecond
	}
	return delay
}

func parseRetryAfter(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(value); err == nil {
		return time.Duration(n) * time.Second, true
	}
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0, true
		}
		return d, true
	}
	return 0, false
}
