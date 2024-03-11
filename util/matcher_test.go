package util

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/stretchr/testify/assert"
)

func TestGlobMultiMatcher(t *testing.T) {
	t.Run("no-pattern-no-match", func(t *testing.T) {
		match, ptn := GlobMultiMatcher()("any")
		assert.False(t, match)
		assert.Empty(t, ptn)
		match, ptn = GlobMultiMatcher(nil...)("any")
		assert.False(t, match)
		assert.Empty(t, ptn)
	})

	t.Run("matches-glob", func(t *testing.T) {
		tests := []struct {
			pattern, value string
			matches        bool
		}{
			{pattern: "", value: "", matches: true},
			{pattern: "*", value: "", matches: true},
			{pattern: "*", value: "value", matches: true},
			{pattern: "", value: "value", matches: false},
			{pattern: "[0-9]", value: "5", matches: true},
			{pattern: "[0-9]", value: "10", matches: false},
			{pattern: "direct", value: "direct", matches: true},
			{pattern: "prefix", value: "prefix-suffix", matches: false},
			{pattern: "suffix", value: "prefix-suffix", matches: false},
			{pattern: "prefix*", value: "prefix-suffix", matches: true},
			{pattern: "*suffix", value: "prefix-suffix", matches: true},
		}
		for i, test := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				match, ptn := GlobMultiMatcher(test.pattern)(test.value)
				assert.Equal(t, match, test.matches, "%q / %q", test.pattern, test.value)
				if match {
					assert.Equal(t, test.pattern, ptn)
				} else {
					assert.Empty(t, ptn)
				}
			})
		}
	})

	t.Run("logs-first-error", func(t *testing.T) {
		defaultLogger := clog.GetDefaultLogger()
		defer clog.SetDefaultLogger(defaultLogger)

		log := clog.NewMemoryHandler()
		clog.SetDefaultLogger(clog.NewLogger(log))

		invalidMatcher := GlobMultiMatcher("longer[", "*[")
		for run := 0; run < 10; run++ {
			invalidMatcher("longer-value")
		}

		assert.Equal(t, "glob matcher (first error is logged): failed matching with *[: syntax error in pattern", log.Pop())
		assert.True(t, log.Empty())
	})

	t.Run("matches-shortest-first", func(t *testing.T) {
		patterns := []string{"-----*", "----*", "---*", "--*", "-*"}
		shuffle := func() {
			rand.Shuffle(len(patterns), func(i, j int) { patterns[i], patterns[j] = patterns[j], patterns[i] })
		}
		for run := 0; run < 1000; run++ {
			shuffle()
			input := slices.Clone(patterns)
			matcher := GlobMultiMatcher(input...)
			for _, pattern := range patterns {
				match, ptn := matcher(pattern)
				assert.True(t, match)
				assert.Equal(t, "-*", ptn)
			}
			assert.Equal(t, input, patterns, "input must not change")
		}
	})

	t.Run("can-be-used-as-condition", func(t *testing.T) {
		matcher := GlobMultiMatcher("a*", "xx", "z*")
		values := []string{"aa", "ab", "ca", "xx", "xy", "z", "z*", "zz"}
		assert.Equal(t, []string{"aa", "ab", "xx", "z", "z*", "zz"}, collect.All(values, matcher.Condition()))
		assert.Equal(t, []string{"aa", "ab", "z", "zz"}, collect.All(values, matcher.NoLiteralMatchCondition()))
	})
}
