package util

import (
	"path"
	"slices"
	"sync"

	"github.com/creativeprojects/clog"
)

// MultiPatternMatcher is a matcher for values using multiple patterns. The first matched pattern is returned
type MultiPatternMatcher func(value string) (match bool, pattern string)

// GlobMultiMatcher returns a function that matches path like values using a set of glob expressions
func GlobMultiMatcher(patterns ...string) MultiPatternMatcher {
	patterns = slices.Clone(patterns)
	slices.Sort(patterns)
	slices.Compact(patterns)                                                    // unique
	slices.SortFunc(patterns, func(a, b string) int { return len(a) - len(b) }) // smallest first

	once := sync.Once{}

	return func(value string) (match bool, pattern string) {
		var err error
		for _, pattern = range patterns {
			match, err = path.Match(pattern, value)
			if err != nil {
				once.Do(func() {
					clog.Warningf("glob matcher (first error is logged): failed matching with %s: %s", pattern, err.Error())
				})
			}
			if match {
				return
			}
		}
		pattern = ""
		return
	}
}
