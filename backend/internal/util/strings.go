package util

import (
	"regexp"
	"strings"
)

var (
	nonAlphanum = regexp.MustCompile(`[^a-z0-9]+`)
	dashTrim    = regexp.MustCompile(`(^-+)|(-+$)`)
)

// Slugify converts an arbitrary string into a URL/profile-safe slug.
// "Harian Pro!" -> "harian-pro".
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanum.ReplaceAllString(s, "-")
	s = dashTrim.ReplaceAllString(s, "")
	return s
}
