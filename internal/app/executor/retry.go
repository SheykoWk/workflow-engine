package executor

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"
)

const maxRetryDelay = 24 * time.Hour

// retryDelayWithJitter returns base * 2^(failedAttempt-1) seconds with ±20% uniform jitter.
// failedAttempt is the attempt number that just failed (1-based).
func retryDelayWithJitter(baseSeconds int, failedAttempt int) time.Duration {
	if baseSeconds < 0 {
		baseSeconds = 0
	}
	if failedAttempt < 1 {
		failedAttempt = 1
	}
	secs := float64(baseSeconds) * math.Pow(2, float64(failedAttempt-1))
	if secs*float64(time.Second) > float64(maxRetryDelay) {
		secs = float64(maxRetryDelay) / float64(time.Second)
	}
	// Uniform jitter in [0.8, 1.2] (~±20%).
	factor := 0.8 + rand.Float64()*0.4
	secs *= factor
	return time.Duration(secs * float64(time.Second))
}

// isNonRetriable reports errors that should not trigger another attempt.
func isNonRetriable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	var hs *HTTPStatusError
	if errors.As(err, &hs) && hs.NonRetriable() {
		return true
	}
	var unk errUnknownStepType
	if errors.As(err, &unk) {
		return true
	}
	var inv errInvalidDelaySeconds
	if errors.As(err, &inv) {
		return true
	}
	return false
}
