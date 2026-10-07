package nativeexport

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWindowsPowerShellCommandsUseNativeModules(t *testing.T) {
	// Deliberately inherit an unusable module path through Go. Both production
	// PowerShell launchers must remove it for the child, without touching Office.
	t.Setenv("PSModulePath", filepath.Join(t.TempDir(), "foreign PowerShell modules"))
	path := filepath.Join(t.TempDir(), "hash input with spaces.txt")
	if err := os.WriteFile(path, []byte("native module hash fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	script := "$h = Get-FileHash -Algorithm SHA256 -LiteralPath '" + strings.ReplaceAll(path, "'", "''") + "'; @{hash=$h.Hash;major=$PSVersionTable.PSVersion.Major} | ConvertTo-Json -Compress"
	for _, run := range []runner{command, roundTripCommand} {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		data, err := run(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
		cancel()
		if err != nil {
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
