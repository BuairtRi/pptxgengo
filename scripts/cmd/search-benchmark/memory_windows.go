package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

var processMemory = windows.NewLazySystemDLL("kernel32.dll").NewProc("K32GetProcessMemoryInfo")

type processMemoryCounters struct {
	Size, PageFaultCount                               uint32
	PeakWorkingSetSize, WorkingSetSize                 uintptr
	QuotaPeakPagedPoolUsage, QuotaPagedPoolUsage       uintptr
	QuotaPeakNonPagedPoolUsage, QuotaNonPagedPoolUsage uintptr
	PagefileUsage, PeakPagefileUsage                   uintptr
}

func peakResidentBytes() (uint64, string, error) {
	var counters processMemoryCounters
	counters.Size = uint32(unsafe.Sizeof(counters))
	ok, _, err := processMemory.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&counters)), uintptr(counters.Size))
	if ok == 0 {
		return 0, "", fmt.Errorf("K32GetProcessMemoryInfo: %w", err)
	}
	return uint64(counters.PeakWorkingSetSize), "Windows K32GetProcessMemoryInfo.PeakWorkingSetSize bytes", nil
}
