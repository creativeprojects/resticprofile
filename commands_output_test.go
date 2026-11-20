package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "absent returns plain", args: []string{"profiles"}, want: "plain"},
		{name: "equals form", args: []string{"profiles", "--output=json"}, want: "json"},
		{name: "space form", args: []string{"profiles", "--output", "json"}, want: "json"},
		{name: "explicit plain", args: []string{"--output=plain"}, want: "plain"},
		{name: "unknown value returned as-is", args: []string{"--output=yaml"}, want: "yaml"},
		{name: "unrelated args ignored", args: []string{"--no-start", "--all", "--reload"}, want: "plain"},
		{name: "last occurrence wins", args: []string{"--output=plain", "--output=json"}, want: "json"},
		{name: "trailing --output without value treated as absent", args: []string{"--output"}, want: "plain"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, parseOutputFormat(tc.args))
		})
	}
}

func TestWantsStructuredOutput(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: nil, want: false},
		{name: "no output flag", args: []string{"profiles", "--all"}, want: false},
		{name: "plain is not structured", args: []string{"--output=plain"}, want: false},
		{name: "json is structured", args: []string{"--output=json"}, want: true},
		{name: "space-separated json", args: []string{"--output", "json"}, want: true},
		{name: "unknown value is treated as structured", args: []string{"--output=yaml"}, want: true},
		{name: "empty value is not structured", args: []string{"--output="}, want: false},
		{name: "trailing --output without value", args: []string{"--output"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, wantsStructuredOutput(tc.args))
		})
	}
}
