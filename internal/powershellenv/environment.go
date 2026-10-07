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

// NativeModuleImports is kept identical to the standalone native/installer scripts.
// These absolute built-in paths bypass unrelated global module discovery.
const NativeModuleImports = `# Load only the required Windows PowerShell modules from this host's PSHOME.
# Explicit paths avoid global module discovery and PowerShell 7/user modules.
Import-Module ([IO.Path]::Combine($PSHOME, 'Modules\Microsoft.PowerShell.Management\Microsoft.PowerShell.Management.psd1')) -ErrorAction Stop
Import-Module ([IO.Path]::Combine($PSHOME, 'Modules\Microsoft.PowerShell.Utility\Microsoft.PowerShell.Utility.psd1')) -ErrorAction Stop
`
