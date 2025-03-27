//go:build darwin

package darwin

import (
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
)

func TestNewProcessType(t *testing.T) {
	tests := []struct {
		name             string
		schedulePriority string
		expected         ProcessType
	}{
		{
			name:             "Background priority",
			schedulePriority: constants.SchedulePriorityBackground,
			expected:         ProcessTypeBackground,
		},
		{
			name:             "Standard priority",
			schedulePriority: constants.SchedulePriorityStandard,
			expected:         ProcessTypeStandard,
		},
		{
			name:             "Unknown priority",
			schedulePriority: "Unknown",
			expected:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewProcessType(tt.schedulePriority)
			if result != tt.expected {
				t.Errorf("NewProcessType(%q) = %q; want %q", tt.schedulePriority, result, tt.expected)
			}
		})
	}
}
