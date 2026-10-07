//go:build !darwin && !linux && !windows

package main

import "fmt"

func peakResidentBytes() (uint64, string, error) {
	return 0, "", fmt.Errorf("process peak memory measurement is unsupported on this OS")
}
