package main

import (
	"errors"
	"math"
	"runtime"
	"strings"
	"testing"
)

func TestElapsedClockUnitsAndInvalidCounters(t *testing.T) {
	for _, frequency := range []int64{1_000, 10_000_000, 1_000_000_000} {
		// Retain precision even when the absolute counter is above float64's
		// exact integer range. The 1ms delta is subtracted as an integer first.
		start := int64(1 << 60)
		c := elapsedClock{method: "fixture", frequency: frequency, read: func() (int64, error) { return start + frequency/1000, nil }}
		got, err := c.since(start)
		if err != nil || got != 1 {
			t.Fatal(frequency, got, err)
		}
	}
	for _, end := range []int64{-1, 0, 10, 9} {
		c := elapsedClock{method: "fixture", frequency: 1000, read: func() (int64, error) { return end, nil }}
		if _, err := c.since(10); err == nil {
			t.Fatal("zero/backward/negative counter accepted", end)
		}
	}
	cause := errors.New("counter failed")
	c := elapsedClock{method: "fixture", frequency: 1000, read: func() (int64, error) { return 0, cause }}
	if _, err := c.sample(); !errors.Is(err, cause) {
		t.Fatal("counter error hidden", err)
	}
	for _, invalid := range []elapsedClock{{}, {method: "fixture", frequency: -1, read: c.read}, {method: "fixture", frequency: 1000}} {
		if _, err := invalid.sample(); err == nil {
			t.Fatal("invalid clock accepted")
		}
	}
}

func TestElapsedClockNativeCounterProgress(t *testing.T) {
	c, err := newElapsedClock()
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		if !strings.Contains(c.method, "QueryPerformanceCounter") || c.frequency < 1_000_000 {
			t.Fatal("native high-resolution Windows counter not selected", c.method, c.frequency)
		}
	} else if c.frequency != 1_000_000_000 || !strings.Contains(c.method, "monotonic") {
		t.Fatal("native Go monotonic counter not selected", c.method, c.frequency)
	}
	first, err := c.sample()
	if err != nil {
		t.Fatal(err)
	}
	previous := first
	for range 1000 {
		next, err := c.sample()
		if err != nil || next < previous {
			t.Fatal("native counter moved backwards", previous, next, err)
		}
		previous = next
	}
	ms, err := c.since(first)
	if err != nil || ms <= 0 || math.IsNaN(ms) || math.IsInf(ms, 0) {
		t.Fatal("native counter did not resolve bounded reads", ms, err)
	}
	t.Logf("elapsed counter: method=%s units_per_second=%d bounded_reads_ms=%g", c.method, c.frequency, ms)
}
