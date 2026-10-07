//go:build !windows

package deckproject

import "testing"

// Unix fixture creation already uses 0700 directories and 0600 key files.
func protectTestIssuer(t *testing.T, path string) { t.Helper() }
