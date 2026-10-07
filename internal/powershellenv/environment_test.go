package powershellenv

import (
	"reflect"
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
