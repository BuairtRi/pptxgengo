package nativeexport

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/buairtri/pptxgengo/internal/powershellenv"
)

func TestWindowsPowerShellCommandsUseNativeModules(t *testing.T) {
	// Deliberately inherit an unusable module path through Go. Both production
	// PowerShell launchers must remove it for the child, without touching Office.
	t.Setenv("PSModulePath", filepath.Join(t.TempDir(), "foreign PowerShell modules"))
	path := filepath.Join(t.TempDir(), "hash input with spaces.txt")
	if err := os.WriteFile(path, []byte("native module hash fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "startup stage.txt")
	quotedMarker := "'" + strings.ReplaceAll(marker, "'", "''") + "'"
	script := "[System.IO.File]::WriteAllText(" + quotedMarker + ", 'started'); $h = Get-FileHash -Algorithm SHA256 -LiteralPath '" + strings.ReplaceAll(path, "'", "''") + "'; [System.IO.File]::WriteAllText(" + quotedMarker + ", 'hashed'); @{hash=$h.Hash;major=$PSVersionTable.PSVersion.Major} | ConvertTo-Json -Compress"
	for _, run := range []runner{command, roundTripCommand} {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		data, err := run(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
		cancel()
		if err != nil {
			resolved, lookupErr := exec.LookPath("powershell.exe")
			stage, stageErr := os.ReadFile(marker)
			t.Logf("Windows PowerShell diagnostic: Go architecture=%s resolved=%q lookup=%v stage=%q stage_error=%v", runtime.GOARCH, resolved, lookupErr, stage, stageErr)
			// Compare the file entry point used by production helpers, with the
			// same child-only environment cleanup and original 30-second bound.
			probe := filepath.Join(t.TempDir(), "bounded startup probe.ps1")
			if writeErr := os.WriteFile(probe, []byte("[Console]::WriteLine($PSVersionTable.PSVersion.ToString()); [Console]::WriteLine([Environment]::Is64BitProcess)"), 0600); writeErr == nil {
				probeCtx, stop := context.WithTimeout(t.Context(), 30*time.Second)
				cmd := exec.CommandContext(probeCtx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", probe)
				cmd.Env = powershellenv.ForWindowsPowerShell(os.Environ())
				cmd.WaitDelay = 2 * time.Second
				started := time.Now()
				probeData, probeErr := cmd.CombinedOutput()
				stop()
				t.Logf("bounded -File startup: elapsed=%s error=%v output=%q", time.Since(started), probeErr, probeData)
			}
			t.Fatalf("native PowerShell modules: %v: %s", err, data)
		}
		var result struct {
			Hash  string
			Major int
		}
		if err := json.Unmarshal(data, &result); err != nil || strings.ToLower(result.Hash) != fmt.Sprintf("%x", sha256.Sum256([]byte("native module hash fixture"))) || result.Major != 5 {
			t.Fatal("wrong Windows PowerShell/hash evidence", string(data), err)
		}
	}
}
