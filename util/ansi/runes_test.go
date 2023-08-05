package ansi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunesLength(t *testing.T) {
	tests := []struct {
		input              []rune
		max, index, length int
	}{
		{input: []rune{}, index: 0, length: 0, max: -1},
		{input: []rune(""), index: 0, length: 0, max: -1},
		{input: []rune(ClearLine + ""), index: len(ClearLine), length: 0, max: -1},
		{input: []rune(ClearLine + "" + ClearLine), index: 2 * len(ClearLine), length: 0, max: -1},
		{input: []rune(ClearLine + "◷" + ClearLine), index: 1 + 2*len(ClearLine), length: 1, max: -1},

		{input: []rune(ClearLine + "◷◷◷" + ClearLine), index: len(ClearLine), length: 3, max: 0},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine), index: 1 + len(ClearLine), length: 3, max: 1},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine), index: 2 + len(ClearLine), length: 3, max: 2},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine), index: 3 + 2*len(ClearLine), length: 3, max: 3},

		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 3 + 2*len(ClearLine), length: 6, max: 3},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 4 + 2*len(ClearLine), length: 6, max: 4},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 5 + 2*len(ClearLine), length: 6, max: 5},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 6 + 2*len(ClearLine), length: 6, max: 6},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 6 + 2*len(ClearLine), length: 6, max: 7},
		{input: []rune(ClearLine + "◷◷◷" + ClearLine + "123"), index: 6 + 2*len(ClearLine), length: 6, max: -1},
	}
	for idx, test := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			length, index := RunesLength(test.input, test.max)
			assert.Equal(t, test.length, length, "length")
			assert.Equal(t, test.index, index, "index")
		})
	}
}
