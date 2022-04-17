package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStack(t *testing.T) {
	trace := getStack(0)
	lines := strings.Split(trace, "\n")
	assert.Greater(t, len(lines), 1)
	assert.Contains(t, lines[0], "resticprofile.getStack")
}
