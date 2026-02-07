// Package testutil provides shared test helpers for EPUB validator tests.
package testutil

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

// HasCode reports whether any diagnostic in diags has the given code.
func HasCode(diags []epub.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

// DiagCodes returns a set of all non-empty diagnostic codes.
func DiagCodes(diags []epub.Diagnostic) map[string]bool {
	codes := make(map[string]bool)
	for _, d := range diags {
		if d.Code != "" {
			codes[d.Code] = true
		}
	}
	return codes
}

// ExpectCode fails the test if the code set does not contain the given code.
func ExpectCode(t *testing.T, codes map[string]bool, code string) {
	t.Helper()
	if !codes[code] {
		t.Errorf("expected diagnostic code %s", code)
	}
}

// SeverityName returns a human-readable name for a diagnostic severity.
func SeverityName(s int) string {
	switch s {
	case epub.SeverityError:
		return "Error"
	case epub.SeverityWarning:
		return "Warning"
	case epub.SeverityInfo:
		return "Info"
	default:
		return "Unknown"
	}
}
