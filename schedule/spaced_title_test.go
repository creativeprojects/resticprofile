//go:build darwin

package schedule

import "testing"

func TestSpacedTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NoSpacesHere", "No Spaces Here"},
		{"Already Spaced", "Already Spaced"},
		{"", ""},
		{"lowercase", "lowercase"},
		{"ALLCAPS", "A L L C A P S"},
		{"iPhone", "i Phone"},
		{"iOS15Device", "i O S15 Device"},
		{"user@Home", "user@ Home"},
	}

	for _, test := range tests {
		result := spacedTitle(test.input)
		if result != test.expected {
			t.Errorf("spacedTitle(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
