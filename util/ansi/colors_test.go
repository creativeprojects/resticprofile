package ansi

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func Test256Colors(t *testing.T) {
	test := func(color *color.Color, expected string) {
		color.EnableColor()
		seq := strings.Split(color.Sprint("||"), "||")[0]
		assert.Equal(t, expected, seq)
	}
	test(New256FgColor(10), Sequence('m', 38, 5, 10))
	test(New256BgColor(10), Sequence('m', 48, 5, 10))
}
