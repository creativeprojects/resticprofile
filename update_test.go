package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type askYesNoTestData struct {
	input         string
	defaultAnswer bool
	expected      bool
}

func TestAskYesNo(t *testing.T) {
	testData := []askYesNoTestData{
		// Empty answer => will follow the defaultAnswer
		{"", true, true},
		{"", false, false},
		{"\n", true, true},
		{"\n", false, false},
		{"\r\n", true, true},
		{"\r\n", false, false},
		// Garbage answer => will always return false
		{"aa", true, false},
		{"aa", false, false},
		{"aa\n", true, false},
		{"aa\n", false, false},
		{"aa\r\n", true, false},
		{"aa\r\n", false, false},
		// Answer yes
		{"y", true, true},
		{"y", false, true},
		{"y\n", true, true},
		{"y\n", false, true},
		{"y\r\n", true, true},
		{"y\r\n", false, true},
		// Full answer yes
		{"yes", true, true},
		{"yes", false, true},
		{"yes\n", true, true},
		{"yes\n", false, true},
		{"yes\r\n", true, true},
		{"yes\r\n", false, true},
		// Answer no
		{"n", true, false},
		{"n", false, false},
		{"n\n", true, false},
		{"n\n", false, false},
		{"n\r\n", true, false},
		{"n\r\n", false, false},
		// Full answer no
		{"no", true, false},
		{"no", false, false},
		{"no\n", true, false},
		{"no\n", false, false},
		{"no\r\n", true, false},
		{"no\r\n", false, false},
	}
	for _, testItem := range testData {
		result := askYesNo(
			bytes.NewBufferString(testItem.input),
			"message",
			testItem.defaultAnswer,
		)
		assert.Equalf(t, testItem.expected, result, "when input was %q", testItem.input)
	}
}
