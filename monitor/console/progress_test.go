package console

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		expected string
		bytes    uint64
	}{
		{expected: "   0 Bytes", bytes: 0},
		{expected: " 500 Bytes", bytes: 500},
		{expected: "1000 Bytes", bytes: 1000},
		{expected: "1.00 KiB", bytes: 1024},
		{expected: "1023 KiB", bytes: 1023 * 1024},
		{expected: "1.00 MiB", bytes: 1024 * 1024},
		{expected: "1.01 MiB", bytes: 10*1024 + 1024*1024},
		{expected: " 512 MiB", bytes: 512 * 1024 * 1024},
		{expected: "64.0 MiB", bytes: 64 * 1024 * 1024},
		{expected: "64.1 MiB", bytes: 100*1024 + 64*1024*1024},
		{expected: "1.00 GiB", bytes: 1024 * 1024 * 1024},
		{expected: "1.00 TiB", bytes: 1024 * 1024 * 1024 * 1024},
		{expected: "2.00 TiB", bytes: 2 * 1024 * 1024 * 1024 * 1024},
		{expected: "1.50 TiB", bytes: 1.5 * 1024 * 1024 * 1024 * 1024},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			assert.Equal(t, test.expected, formatBytes(test.bytes))
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{expected: "00:00", duration: 0},
		{expected: "00:01", duration: time.Second},
		{expected: "01:00", duration: time.Minute},
		{expected: "01:01", duration: time.Minute + time.Second},
		{expected: "59:59", duration: 59*time.Minute + 59*time.Second},
		{expected: "01:00:00", duration: time.Hour},
		{expected: "01:01:00", duration: time.Hour + time.Minute},
		{expected: "01:01:01", duration: time.Hour + time.Minute + time.Second},
		{expected: "23:59:00", duration: 23*time.Hour + 59*time.Minute},
		{expected: "23:59:59", duration: 24*time.Hour - 1},
		{expected: "1d 00:00:00", duration: 24 * time.Hour},
		{expected: "1d 01:00:00", duration: 25 * time.Hour},
		{expected: "2d 00:00:00", duration: 48 * time.Hour},
		{expected: "2d 01:00:00", duration: 49 * time.Hour},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			assert.Equal(t, test.expected, formatDuration(test.duration))
		})
	}
}

func TestFormatProgress(t *testing.T) {
	tests := []struct {
		style, width int
		progress     float64
		expected     string
	}{
		{expected: "        ", style: 0, width: 8, progress: 0},
		{expected: "##      ", style: 0, width: 8, progress: 0.25},
		{expected: "####    ", style: 0, width: 8, progress: 0.5},
		{expected: "######  ", style: 0, width: 8, progress: 0.75},
		{expected: "########", style: 0, width: 8, progress: 1},

		{expected: "  ", style: 0, width: 2, progress: 1 * 1. / 9},
		{expected: "- ", style: 0, width: 2, progress: 2 * 1. / 9},
		{expected: "+ ", style: 0, width: 2, progress: 3 * 1. / 9},
		{expected: "* ", style: 0, width: 2, progress: 4 * 1. / 9},
		{expected: "# ", style: 0, width: 2, progress: 5 * 1. / 9},
		{expected: "#-", style: 0, width: 2, progress: 6 * 1. / 9},
		{expected: "#+", style: 0, width: 2, progress: 7 * 1. / 9},
		{expected: "#*", style: 0, width: 2, progress: 8 * 1. / 9},
		{expected: "##", style: 0, width: 2, progress: 9 * 1. / 9},

		{expected: "  ", style: 0, width: 2, progress: 0.5 * 1. / 9},
		{expected: "  ", style: 0, width: 2, progress: 1.0 * 1. / 9},
		{expected: "- ", style: 0, width: 2, progress: 1.5 * 1. / 9},
		{expected: "- ", style: 0, width: 2, progress: 2.0 * 1. / 9},
		{expected: "+ ", style: 0, width: 2, progress: 2.5 * 1. / 9},

		{expected: "⡇", style: defaultBarStyle, width: 1, progress: 0.5},
		{expected: "⣿ ", style: defaultBarStyle, width: 2, progress: 0.5},
		{expected: "⣿⡇ ", style: defaultBarStyle, width: 3, progress: 0.5},
		{expected: "⣿⣿  ", style: defaultBarStyle, width: 4, progress: 0.5},
		{expected: "⠿⠿⠿    50.0%", style: defaultBarStyle, width: 12, progress: 0.5},

		{expected: "█▒     25.0%", style: 1, width: 12, progress: 0.25},
		{expected: "███    50.0%", style: 1, width: 12, progress: 0.5},
		{expected: "██████ 99.5%", style: 1, width: 12, progress: 0.995},
		{expected: "██████  100%", style: 1, width: 12, progress: 1},

		{expected: " 50%", style: -1, width: 4, progress: 0.5},
		{expected: "100%", style: -1, width: 4, progress: 1},

		{expected: "50.0%", style: -1, width: 5, progress: 0.5},
		{expected: " 100%", style: -1, width: 5, progress: 1},
		{expected: "50.0%", style: -1, width: 10, progress: 0.5},
		{expected: " 100%", style: -1, width: 10, progress: 1},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			actual := formatProgress(test.style, test.width, false, test.progress)
			assert.Equalf(t, test.expected, actual, "progress: %.2f", test.progress)
		})
	}
}
