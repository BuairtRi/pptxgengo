package main

import "golang.org/x/sys/unix"

func peakResidentBytes() (uint64, string, error) {
	var r unix.Rusage
	err := unix.Getrusage(unix.RUSAGE_SELF, &r)
	// The current Darwin getrusage(2) interface reports ru_maxrss in bytes.
	return uint64(r.Maxrss), "Darwin getrusage(RUSAGE_SELF).ru_maxrss bytes", err
}
