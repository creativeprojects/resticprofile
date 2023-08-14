package main

import (
	"errors"
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test error")

func TestNotRunningOnBattery(t *testing.T) {
	battery, charge, err := IsRunningOnBattery()
	assert.NoError(t, err)
	assert.False(t, battery)
	assert.Zero(t, charge)
}

func TestIsFatalError(t *testing.T) {
	fixtures := []struct {
		err      error
		expected error
	}{
		{
			nil,
			nil,
		},
		{
			errTest,
			errTest,
		},
		{
			battery.ErrFatal{},
			battery.ErrFatal{},
		},
		{
			battery.ErrPartial{
				State: nil,
			},
			nil,
		},
		{
			battery.ErrPartial{
				State: errTest,
			},
			errTest,
		},
		{
			battery.Errors{nil},
			nil,
		},
		{
			battery.Errors{nil, errTest},
			errTest,
		},
		{
			battery.Errors{nil, battery.ErrPartial{State: errTest}},
			errTest,
		},
		{
			battery.Errors{nil, battery.ErrPartial{State: nil}},
			nil,
		},
		{
			battery.Errors{
				battery.ErrPartial{State: nil},
				battery.ErrPartial{State: errTest},
			},
			errTest,
		},
	}

	for _, fixture := range fixtures {
		t.Run("", func(t *testing.T) {
			err := isFatalError(fixture.err)
			assert.Equal(t, fixture.expected, err)
		})
	}
}

func TestDetectRunningOnBattery(t *testing.T) {
	fixtures := []struct {
		batteries        []*battery.Battery
		runningOnBattery bool
	}{
		// one discharging
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
			},
			true,
		},
		// one charging
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
			},
			false,
		},
		// one full
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Full,
					},
				},
			},
			false,
		},
		// one idle (usually at around 80%)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Idle,
					},
				},
			},
			false,
		},
		// one unknown (we consider this as not running on battery)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Unknown,
					},
				},
			},
			false,
		},
		// one discharging, one charging
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
			},
			false,
		},
		// one discharging, one full
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Full,
					},
				},
			},
			false,
		},
		// one discharging, one idle (usually at around 80%)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Idle,
					},
				},
			},
			false,
		},
		// one discharging, one unknown (we consider this as not running on battery)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Unknown,
					},
				},
			},
			false,
		},
		// one charging, one discharging
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
			},
			false,
		},
		// one charging, one full
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Full,
					},
				},
			},
			false,
		},
		// one charging, one idle (usually at around 80%)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Idle,
					},
				},
			},
			false,
		},
		// one charging, one unknown (we consider this as not running on battery)
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Charging,
					},
				},
				{
					State: battery.State{
						Raw: battery.Unknown,
					},
				},
			},
			false,
		},
		// one full, one discharging
		{
			[]*battery.Battery{
				{
					State: battery.State{
						Raw: battery.Full,
					},
				},
				{
					State: battery.State{
						Raw: battery.Discharging,
					},
				},
			},
			false,
		},
	}

	for _, fixture := range fixtures {
		t.Run("", func(t *testing.T) {
			bat := detectRunningOnBattery(fixture.batteries)
			assert.Equal(t, fixture.runningOnBattery, bat)
		})
	}
}

func TestAverageBatteryLevel(t *testing.T) {
	fixtures := []struct {
		batteries []*battery.Battery
		average   float64
	}{
		{
			[]*battery.Battery{},
			0,
		},
		{
			[]*battery.Battery{
				{
					Current: 0,
					Full:    0,
				},
			},
			0,
		},
		{
			[]*battery.Battery{
				{
					Current: 10,
					Full:    100,
				},
			},
			10,
		},
		{
			[]*battery.Battery{
				{
					Current: 10,
					Full:    100,
				},
				{
					Current: 20,
					Full:    100,
				},
			},
			15,
		},
	}

	for _, fixture := range fixtures {
		t.Run("", func(t *testing.T) {
			bat := averageBatteryLevel(fixture.batteries)
			assert.Equal(t, fixture.average, bat)
		})
	}
}
