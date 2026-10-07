// Package powershellenv prepares child environments for Windows PowerShell 5.1.
package powershellenv

import "strings"

// ForWindowsPowerShell lets powershell.exe build its own native module search
// path. A Go process launched by pwsh otherwise passes PowerShell 7 module paths
// to its grandchild, breaking built-in Utility cmdlets such as Get-FileHash and
// ConvertTo-Json. Only the child environment changes; user/registry paths remain.
// https://learn.microsoft.com/powershell/module/microsoft.powershell.core/about/about_psmodulepath
func ForWindowsPowerShell(base []string) []string {
	out := make([]string, 0, len(base))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, "PSModulePath") {
			out = append(out, entry)
		}
	}
	return out
}
