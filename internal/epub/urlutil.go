package epub

import (
	"slices"
	"strings"
)

// IsRemoteURL reports whether href starts with http:// or https://.
func IsRemoteURL(href string) bool {
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

// StripFragment removes the fragment (everything from # onward) from ref.
// If there is no fragment, ref is returned unchanged.
func StripFragment(ref string) string {
	if before, _, ok := strings.Cut(ref, "#"); ok {
		return before
	}
	return ref
}

// ContainsToken reports whether a space-separated token list contains token.
func ContainsToken(tokenList, token string) bool {
	return slices.Contains(strings.Fields(tokenList), token)
}
