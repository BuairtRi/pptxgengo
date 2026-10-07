package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func peakResidentBytes() (uint64, string, error) {
	// getrusage retains the pre-exec address-space high-water mark. Go's
	// fork/exec can therefore carry model preparation memory from the parent.
	// VmHWM belongs to this executable's current mm, rather than signal.maxrss.
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, "", err
	}
	value, err := linuxPeakFromStatus(string(data))
	return value, "Linux /proc/self/status VmHWM KiB converted to bytes (current executable address space)", err
}

func linuxPeakFromStatus(status string) (uint64, error) {
	for _, line := range strings.Split(status, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "VmHWM:" {
			continue
		}
		if len(fields) != 3 || fields[2] != "kB" {
			return 0, fmt.Errorf("invalid Linux VmHWM units")
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || v == 0 || v > ^uint64(0)/1024 {
			return 0, fmt.Errorf("invalid Linux VmHWM value")
		}
		return v * 1024, nil
	}
	return 0, fmt.Errorf("Linux VmHWM unavailable")
}
