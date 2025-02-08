package stringutil

import "strings"

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
