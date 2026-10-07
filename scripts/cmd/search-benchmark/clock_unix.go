//go:build !windows

package main

import "time"

func newElapsedClock() (elapsedClock, error) {
	origin := time.Now()
	return elapsedClock{
		method:    "Go time.Since monotonic nanosecond counter",
		frequency: 1_000_000_000,
		read:      func() (int64, error) { return time.Since(origin).Nanoseconds(), nil },
	}, nil
}
