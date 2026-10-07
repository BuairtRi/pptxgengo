//go:build !windows

package installstate

import "testing"

func qualifyWindowsInstallerScript(t *testing.T, repo, work, binaries string) []string {
	t.Helper()
	t.Fatal("Windows installer qualification called on a non-Windows platform")
	return nil
}
