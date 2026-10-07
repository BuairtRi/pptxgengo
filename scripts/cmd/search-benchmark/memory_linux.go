package main

import "golang.org/x/sys/unix"

func peakResidentBytes() (uint64, string, error) {
	var r unix.Rusage
	err := unix.Getrusage(unix.RUSAGE_SELF, &r)
	return uint64(r.Maxrss) * 1024, "Linux getrusage(RUSAGE_SELF).ru_maxrss KiB converted to bytes", err
}
