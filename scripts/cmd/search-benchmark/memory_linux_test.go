package main

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestLinuxPeakStatusParsing(t *testing.T) {
	value, err := linuxPeakFromStatus("Name:\tbenchmark\nVmHWM:\t12345 kB\nVmRSS: 10 kB\n")
	if err != nil || value != 12345*1024 {
		t.Fatal(value, err)
	}
	for _, invalid := range []string{"", "VmHWM: 0 kB", "VmHWM: -1 kB", "VmHWM: 12 bytes", "VmHWM: 12", "VmHWM: 18446744073709551615 kB"} {
		if _, err := linuxPeakFromStatus(invalid); err == nil {
			t.Fatal("invalid Linux memory evidence accepted", invalid)
		}
	}
}

func TestLinuxPeakExcludesPreExecParent(t *testing.T) {
	const helper = "PPTXGENGO_BENCH_MEMORY_HELPER"
	if os.Getenv(helper) == "1" {
		peak, _, err := peakResidentBytes()
		if err != nil {
			t.Fatal(err)
		}
		// A raw value is easier to check than the Go test runner's output.
		if err := os.WriteFile(os.Getenv(helper+"_OUT"), []byte(strconv.FormatUint(peak, 10)), 0600); err != nil {
			t.Fatal(err)
		}
		return
	}
	pages := make([]byte, 192<<20)
	for i := 0; i < len(pages); i += 4096 {
		pages[i] = 1
	}
	parentPeak, _, err := peakResidentBytes()
	if err != nil || parentPeak < uint64(len(pages)) {
		t.Fatal("parent memory was not resident", parentPeak, err)
	}
	out := t.TempDir() + "/child-peak"
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), self, "-test.run=^TestLinuxPeakExcludesPreExecParent$", "-test.count=1")
	cmd.Env = append(os.Environ(), helper+"=1", helper+"_OUT="+out)
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("memory helper: %v: %s", err, data)
	}
	runtime.KeepAlive(pages)
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	childPeak, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || childPeak == 0 || childPeak >= uint64(len(pages)) {
		t.Fatal("child peak includes pre-exec parent memory", childPeak, parentPeak, err)
	}
}
