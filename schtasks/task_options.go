//go:build windows

package schtasks

import "time"

type TaskOption interface {
	apply(*Task)
}

type WithFromNowOption struct {
	now time.Time
}

func WithFromNow(now time.Time) WithFromNowOption {
	return WithFromNowOption{now: now}
}

func (w WithFromNowOption) apply(t *Task) {
	t.setFromNow(w.now)
}
