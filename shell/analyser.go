package shell

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/monitor"
)

type analyserCallback struct {
	requiredCount,
	invoked,
	lastInvocation,
	maxInvocations int
	stopOnError bool
	fn          func(line string) error
}

type analyserPattern struct {
	name       string
	expression *regexp.Regexp
}

type OutputAnalyser struct {
	lock      sync.Locker
	patterns  []analyserPattern
	callbacks map[string]*analyserCallback
	counts    map[string]int
	matches   map[string][]string
}

var outputAnalyserPatterns = map[string]*regexp.Regexp{
	"lock-failure,who":    regexp.MustCompile("unable to create lock.+already locked.+?by (.+)$"),
	"lock-failure,age":    regexp.MustCompile("lock was created at.+\\(([^()]+)\\s+ago\\)"),
	"lock-failure,stale":  regexp.MustCompile("the\\W+unlock\\W+command can be used to remove stale locks"),
	"lock-retry,max-wait": regexp.MustCompile("repo already locked, waiting up to (\\S+) for the lock"),
}

func NewOutputAnalyser() *OutputAnalyser {
	patterns := make([]analyserPattern, 0, len(outputAnalyserPatterns))
	for name, regex := range outputAnalyserPatterns {
		patterns = append(patterns, analyserPattern{name: name, expression: regex})
	}

	return (&OutputAnalyser{
		lock:      &sync.Mutex{},
		patterns:  patterns,
		callbacks: map[string]*analyserCallback{},
	}).Reset()
}

func (a *OutputAnalyser) Reset() *OutputAnalyser {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.counts = map[string]int{}
	a.matches = map[string][]string{}
	return a
}

// SetCallback registers a custom callback that is invoked when pattern (regex) matches a line.
func (a *OutputAnalyser) SetCallback(name, pattern string, minCount, maxCalls int, stopOnError bool, callback func(line string) error) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	name = fmt.Sprintf("#%s", strings.TrimSpace(strings.Split(name, ",")[0]))

	// Removing existing pattern
	for i := len(a.patterns) - 1; i >= 0; i-- {
		if a.patterns[i].name == name {
			a.patterns = append(a.patterns[0:i], a.patterns[i+1:]...)
		}
	}

	if callback == nil {
		delete(a.callbacks, name)
	} else {
		a.patterns = append(a.patterns, analyserPattern{
			name:       name,
			expression: regex,
		})
		a.callbacks[name] = &analyserCallback{
			requiredCount:  minCount,
			maxInvocations: maxCalls,
			stopOnError:    stopOnError,
			fn:             callback,
		}
	}
	return nil
}

func (a *OutputAnalyser) invokeCallback(name, line string) (err error) {
	count, hasCount := a.counts[name]
	cb, hasCallback := a.callbacks[name]

	if hasCount && hasCallback {
		minCount := cb.requiredCount + cb.lastInvocation
		if count >= minCount && (cb.invoked < cb.maxInvocations || cb.maxInvocations <= 0) {
			cb.lastInvocation = count
			cb.invoked++
			err = cb.fn(line)
		}

		if err != nil {
			if cb.stopOnError {
				clog.Errorf("stream-error action '%s' failed: %s", name, err.Error())
			} else {
				clog.Warningf("stream-error action '%s' failed: %s", name, err.Error())
				err = nil
			}
		}
	}

	return
}

func (a *OutputAnalyser) ContainsRemoteLockFailure() bool {
	a.lock.Lock()
	defer a.lock.Unlock()

	if count, ok := a.counts["lock-failure"]; ok {
		return count >= 2
	}
	return false
}

func (a *OutputAnalyser) GetRemoteLockedSince() (time.Duration, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if m, ok := a.matches["lock-failure,age"]; ok && len(m) > 1 {
		if lockAge, err := time.ParseDuration(m[1]); err == nil {
			return lockAge.Truncate(time.Millisecond), true
		} else {
			clog.Warningf("failed parsing restic lock age. Cause %s", err.Error())
		}
	}

	return 0, false
}

func (a *OutputAnalyser) GetRemoteLockedMaxWait() (time.Duration, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if m, ok := a.matches["lock-retry,max-wait"]; ok && len(m) > 1 {
		if maxWait, err := time.ParseDuration(m[1]); err == nil {
			return maxWait.Truncate(time.Millisecond), true
		} else {
			clog.Debugf("failed parsing restic retry-lock max wait. Cause %s", err.Error())
		}
	}

	return 0, false
}

func (a *OutputAnalyser) GetRemoteLockedBy() (string, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if m, ok := a.matches["lock-failure,who"]; ok && len(m) > 1 {
		return m[1], true
	}

	return "", false
}

func (a *OutputAnalyser) AnalyseStringLines(output string) error {
	return a.AnalyseLines(strings.NewReader(output))
}

func (a *OutputAnalyser) AnalyseLines(output io.Reader) (err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		if err == nil {
			if err = a.analyseLine(scanner.Text()); err != nil {
				clog.Warningf("output analysis stopped after error")
			}
		}
	}

	if err == nil {
		err = scanner.Err()
	}
	return
}

func (a *OutputAnalyser) analyseLine(line string) (err error) {
	for _, pattern := range a.patterns {
		match := pattern.expression.FindStringSubmatch(line)
		if match != nil {
			a.matches[pattern.name] = match

			baseName := strings.Split(pattern.name, ",")[0]
			a.counts[baseName]++

			if err == nil {
				err = a.invokeCallback(baseName, line)
			}
		}
	}
	return
}

// Ensure OutputAnalyser implements OutputAnalysis
var _ = monitor.OutputAnalysis(NewOutputAnalyser())
