package releasepackage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWindowsInstallerPowerShellSyntax(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell parser requires Windows")
	}
	path := filepath.Join(t.TempDir(), "installer with spaces.ps1")
	if e := os.WriteFile(path, installer, 0600); e != nil {
		t.Fatal(e)
	}
	literal := "'" + strings.ReplaceAll(path, "'", "''") + "'"
	command := "$tokens=$null; $parseErrors=$null; [void][System.Management.Automation.Language.Parser]::ParseFile(" + literal + ",[ref]$tokens,[ref]$parseErrors); if ($parseErrors.Count -gt 0) { $parseErrors | ForEach-Object { $_.Message }; exit 1 }"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if out, e := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", command).CombinedOutput(); e != nil {
		t.Fatalf("PowerShell parser: %v: %s", e, out)
	}
}
