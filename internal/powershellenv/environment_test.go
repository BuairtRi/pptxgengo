package powershellenv

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWindowsPowerShellChildEnvironment(t *testing.T) {
	base := []string{"Path=C:\\tools", "PSModulePath=C:\\PowerShell7\\Modules", "psmodulepath=another", "WinPSModulePath=native", "PSModulePathExtra=preserved", "OTHER=value=with=equals"}
	before := append([]string{}, base...)
	want := []string{base[0], base[3], base[4], base[5]}
	got := ForWindowsPowerShell(base)
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(base, before) {
		t.Fatal("unexpected environment mutation", got, base)
	}
	got[0] = "changed child entry"
	if base[0] != before[0] {
		t.Fatal("child environment aliases parent slice")
	}
}

func TestStandaloneScriptsUseNativeModuleImports(t *testing.T) {
	for _, name := range []string{"internal/nativeexport/powerpoint.ps1", "internal/nativeexport/roundtrip.ps1", "internal/releasepackage/install-windows.ps1", "internal/releasepackage/smoke-test-windows.ps1"} {
		data, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), NativeModuleImports) != 1 {
			t.Fatalf("%s: native import block drifted", name)
		}
		if strings.Index(string(data), NativeModuleImports) > strings.Index(string(data), "New-Object") {
			t.Fatalf("%s: imports must precede Utility cmdlets", name)
		}
	}
}
