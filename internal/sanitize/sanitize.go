package sanitize

import (
	"regexp"
	"strings"
)

var (
	consecutiveWhitespace = regexp.MustCompile(`(\s{2,})`)
)

// TODO: optimize this later
func Normalize(s string) string {
	return consecutiveWhitespace.ReplaceAllLiteralString(
		strings.TrimSpace(
			strings.ToLower(s),
		),
		" ",
	)
}
