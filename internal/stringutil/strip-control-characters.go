package stringutil

import (
	"regexp"
	"strings"
)

// StripControlCharacters removes ASCII control characters from the input string.
func StripControlCharacters(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if r >= 32 && r != 127 { // Keep printable characters (32-126)
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// StripColorCodes removes ANSI color codes from a string.
func StripColorCodes(s string) string {
	// Regular expression to match ANSI escape codes
	re := regexp.MustCompile(`\x1B\[[0-?9;]*[mK]`)
	return re.ReplaceAllString(s, "")
}
