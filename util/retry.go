package util

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Stop indicates manually stopping of custom retry
type Stop struct {
	error
}

// NewStop function can be used to implement a custom stop message
func NewStop(message string) Stop {
	return Stop{errors.New(message)}
}

// CustomRetry retries a function f for attempts time with given duration
func CustomRetry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(Stop); ok {
			// Return the original error for later checking
			return s.error
		}
		if attempts--; attempts > 0 {
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2

			time.Sleep(sleep)
			return CustomRetry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}
