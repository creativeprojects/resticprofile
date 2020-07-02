package calendar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywords(t *testing.T) {
	testData := []struct{ keyword, expected string }{
		{"minutely", "*-*-* *:*:00"},
		{"hourly", "*-*-* *:00:00"},
		{"daily", "*-*-* 00:00:00"},
		{"monthly", "*-*-01 00:00:00"},
		{"weekly", "Mon *-*-* 00:00:00"},
		{"yearly", "*-01-01 00:00:00"},
		{"quarterly", "*-01,04,07,10-01 00:00:00"},
		{"semiannually", "*-01,07-01 00:00:00"},
	}

	for _, testItem := range testData {
		output := specialKeywords[testItem.keyword].String()
		assert.Equal(t, testItem.expected, output)
	}
}
