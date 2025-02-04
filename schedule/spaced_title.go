//go:build darwin

package schedule

import "strings"

// spacedTitle adds spaces before capital letters in a string, except for the first character
// or when the capital letter follows a space. For example, "ThisIsATest" becomes "This Is A Test".
// If the input is empty or contains no capital letters, it is returned unchanged.
// This function is only used by the launchd handler on macOS.
//
// Example:
//
//	spacedTitle("ResticProfile") // Returns "Restic Profile"
//	spacedTitle("ABC")          // Returns "A B C"
//	spacedTitle("already spaced") // Returns "already spaced"
func spacedTitle(title string) string {
	if title == "" {
		return title
	}
	var previous rune
	sb := strings.Builder{}
	for _, char := range title {
		if char >= 'A' && char <= 'Z' && previous != ' ' && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteRune(char)
		previous = char
	}
	return sb.String()
}
