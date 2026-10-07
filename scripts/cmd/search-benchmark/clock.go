package main

import "fmt"

// Counter units describe representation, not a promise of physical accuracy.
// Subtract integer samples before converting to avoid precision loss at uptime.
type elapsedClock struct {
	method    string
	frequency int64
	read      func() (int64, error)
}

func (c elapsedClock) sample() (int64, error) {
	if c.frequency <= 0 || c.method == "" || c.read == nil {
		return 0, fmt.Errorf("invalid elapsed clock")
	}
	v, err := c.read()
	if err != nil {
		return 0, fmt.Errorf("elapsed counter: %w", err)
	}
	if v < 0 {
		return 0, fmt.Errorf("negative elapsed counter")
	}
	return v, nil
}

func (c elapsedClock) since(start int64) (float64, error) {
	end, err := c.sample()
	if err != nil {
		return 0, err
	}
	if start < 0 || end <= start {
		return 0, fmt.Errorf("elapsed counter did not advance; refuse zero/backward timing")
	}
	return float64(end-start) * 1000 / float64(c.frequency), nil
}
