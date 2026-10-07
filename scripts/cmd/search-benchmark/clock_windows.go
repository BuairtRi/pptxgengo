package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var performanceCounter = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryPerformanceCounter")
var performanceFrequency = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryPerformanceFrequency")

// Go's Windows ARM64 time.Now reads interrupt time, which can quantize a fast
// query to zero. Use the supported monotonic interval counter on both Windows
// architectures, without changing system timer resolution or requesting sleep.
// https://learn.microsoft.com/windows/win32/sysinfo/acquiring-high-resolution-time-stamps
func newElapsedClock() (elapsedClock, error) {
	if err := performanceCounter.Find(); err != nil {
		return elapsedClock{}, err
	}
	if err := performanceFrequency.Find(); err != nil {
		return elapsedClock{}, err
	}
	var frequency int64
	ok, _, err := performanceFrequency.Call(uintptr(unsafe.Pointer(&frequency)))
	if ok == 0 || frequency <= 0 {
		return elapsedClock{}, fmt.Errorf("QueryPerformanceFrequency failed: frequency=%d error=%v", frequency, err)
	}
	return elapsedClock{
		method:    "Windows QueryPerformanceCounter; QueryPerformanceFrequency units per second",
		frequency: frequency,
		read: func() (int64, error) {
			var counter int64
			ok, _, err := performanceCounter.Call(uintptr(unsafe.Pointer(&counter)))
			if ok == 0 {
				return 0, fmt.Errorf("QueryPerformanceCounter: %v", err)
			}
			return counter, nil
		},
	}, nil
}
