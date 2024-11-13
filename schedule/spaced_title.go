//go:build darwin

package schedule

import "strings"

func spacedTitle(title string) string {
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
