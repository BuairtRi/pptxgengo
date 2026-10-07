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
	// Both production launchers must remove this unusable inherited module
	// path. The standalone scripts use this exact native import block, too.
	t.Setenv("PSModulePath", filepath.Join(t.TempDir(), "foreign PowerShell modules"))
	path := filepath.Join(t.TempDir(), "hash input with spaces.txt")
	if err := os.WriteFile(path, []byte("native module hash fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, helper := range []struct {
		name string
		run  runner
	}{{"export", command}, {"roundtrip", roundTripCommand}} {
		for _, entry := range []string{"Command", "File"} {
			t.Run(helper.name+"/"+entry, func(t *testing.T) {
				marker := filepath.Join(t.TempDir(), "script stage.txt")
				quotedMarker := "'" + strings.ReplaceAll(marker, "'", "''") + "'"
				script := "$ErrorActionPreference = 'Stop'; [IO.File]::WriteAllText(" + quotedMarker + ", 'started')\n" + powershellenv.NativeModuleImports +
					"[IO.File]::WriteAllText(" + quotedMarker + ", 'imported'); $h = Get-FileHash -Algorithm SHA256 -LiteralPath '" + strings.ReplaceAll(path, "'", "''") + "'; [IO.File]::WriteAllText(" + quotedMarker + ", 'hashed'); " +
					"$json = @{hash=$h.Hash;major=$PSVersionTable.PSVersion.Major;home=$PSHOME;utility=(Get-Module Microsoft.PowerShell.Utility).Path;management=(Get-Module Microsoft.PowerShell.Management).Path} | ConvertTo-Json -Compress; [IO.File]::WriteAllText(" + quotedMarker + ", 'serialized'); [Console]::WriteLine($json)"
				args := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass"}
				if entry == "File" {
					file := filepath.Join(t.TempDir(), "native modules with spaces.ps1")
					if err := os.WriteFile(file, []byte(script), 0600); err != nil {
						t.Fatal(err)
					}
					args = append(args, "-File", file)
				} else {
					args = append(args, "-Command", script)
				}
				ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
				started := time.Now()
				data, err := helper.run(ctx, "powershell.exe", args...)
				cancel()
				t.Logf("native modules: architecture=%s entry=%s elapsed=%s", runtime.GOARCH, entry, time.Since(started))
				if err != nil {
					resolved, lookupErr := exec.LookPath("powershell.exe")
					stage, stageErr := os.ReadFile(marker)
					t.Fatalf("native PowerShell modules: %v: %s; resolved=%q lookup=%v stage=%q stage_error=%v", err, data, resolved, lookupErr, stage, stageErr)
				}
				var result struct {
					Hash, Home, Utility, Management string
					Major                           int
				}
				if err := json.Unmarshal(data, &result); err != nil || strings.ToLower(result.Hash) != fmt.Sprintf("%x", sha256.Sum256([]byte("native module hash fixture"))) || result.Major != 5 {
					t.Fatal("wrong Windows PowerShell/hash evidence", string(data), err)
				}
				for module, got := range map[string]string{"Microsoft.PowerShell.Utility": result.Utility, "Microsoft.PowerShell.Management": result.Management} {
					want := filepath.Join(result.Home, "Modules", module, module+".psd1")
					if !filepath.IsAbs(result.Home) || !strings.EqualFold(filepath.Clean(got), want) {
						t.Fatalf("module %s loaded outside native PSHOME: got=%q want=%q", module, got, want)
					}
				}
			})
		}
	}
}
