package shell

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
)

type OutputAnalyser struct {
	lock    sync.Locker
	counts  map[string]int
	matches map[string][]string
}

var outputAnalyzerPatterns = map[string]*regexp.Regexp{
	"lock-failure,who":   regexp.MustCompile("unable to create lock.+already locked.+?by (.+)$"),
	"lock-failure,age":   regexp.MustCompile("lock was created at.+\\(([^()]+)\\s+ago\\)"),
	"lock-failure,stale": regexp.MustCompile("the\\W+unlock\\W+command can be used to remove stale locks"),
}

func NewOutputAnalyser() *OutputAnalyser {
	return &OutputAnalyser{
		counts:  map[string]int{},
		matches: map[string][]string{},
		lock:    &sync.Mutex{},
	}
}

func (a OutputAnalyser) ContainsRemoteLockFailure() bool {
	a.lock.Lock()
	defer a.lock.Unlock()

	if count, ok := a.counts["lock-failure"]; ok {
		return count >= 2
	}
	return false
}

func (a OutputAnalyser) GetRemoteLockedSince() (time.Duration, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if m, ok := a.matches["lock-failure,age"]; ok && len(m) > 1 {
		if lockAge, err := time.ParseDuration(m[1]); err == nil {
			return lockAge.Truncate(time.Millisecond), true
		} else {
			clog.Warningf("Failed parsing restic lock age. Cause %s", err.Error())
		}
	}

	return 0, false
}

func (a OutputAnalyser) GetRemoteLockedBy() (string, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if m, ok := a.matches["lock-failure,who"]; ok && len(m) > 1 {
		return m[1], true
	}

	return "", false
}

func (a OutputAnalyser) AnalyzeStringLines(output string) OutputAnalysis {
	return a.AnalyzeLines(strings.NewReader(output))
}

func (a OutputAnalyser) AnalyzeLines(output io.Reader) OutputAnalysis {
	a.lock.Lock()
	defer a.lock.Unlock()

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		a.analyzeLine(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		clog.Warningf("Failed analyzing all output lines. Cause %s", err.Error())
	}

	return a
}

func (a OutputAnalyser) analyzeLine(line string) {
	for name, pattern := range outputAnalyzerPatterns {
		match := pattern.FindStringSubmatch(line)
		if match != nil {
			a.matches[name] = match

			baseName := strings.Split(name, ",")[0]
			a.counts[baseName]++
		}
	}
}

// Ensure OutputAnalyzer implements OutputAnalysis
var _ = OutputAnalysis(NewOutputAnalyser())
