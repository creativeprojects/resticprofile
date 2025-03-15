package batt

import (
	"math"

	"github.com/creativeprojects/clog"
	"github.com/distatus/battery"
)

// Batteries return all connected batteries information
func Batteries() ([]battery.Battery, error) {
	batteries, err := battery.GetAll()
	output := make([]battery.Battery, 0, len(batteries))
	for _, battery := range batteries {
		if battery == nil {
			continue
		}
		if battery.Design == 0 && battery.Full == 0 && battery.Voltage == 0 {
			// bug in recent mac OS hardware that returns ghost battery information
			// https://github.com/distatus/battery/issues/34
			continue
		}
		output = append(output, *battery)
	}
	return output, err
}

// IsRunningOnBattery returns true if the computer is running on battery,
// followed by the percentage of battery remaining. If no battery is present
// the percentage will be 0.
func IsRunningOnBattery() (bool, int, error) {
	batteries, err := Batteries()
	if err != nil {
		err = isFatalError(err)
		if err != nil {
			return false, 0, err
		}
	}
	if len(batteries) == 0 {
		clog.Debug("no battery detected")
		return false, 0, nil
	}
	return detectRunningOnBattery(batteries),
		int(math.Round(averageBatteryLevel(batteries))), nil
}

// isFatalError returns an error if we can't detect any battery state
func isFatalError(err error) error {
	switch perr := err.(type) {

	case battery.ErrFatal: // complete failure
		return err

	case battery.Errors: // range of errors per battery
		for _, err := range perr {
			if err != nil {
				err = isFatalError(err)
				if err != nil {
					return err
				}
			}
		}
		return nil

	case battery.ErrPartial: // only some errors on one battery
		if perr.State != nil {
			return perr.State
		}
		// we're only interested in the state
		return nil
	}

	return err
}

// detectRunningOnBattery returns true if all batteries are discharging
func detectRunningOnBattery(batteries []battery.Battery) bool {
	pluggedIn := false
	discharging := false
	for _, bat := range batteries {
		if bat.State.Raw == battery.Discharging || bat.State.Raw == battery.Empty {
			discharging = true
		} else if bat.State.Raw == battery.Charging || bat.State.Raw == battery.Full || bat.State.Raw == battery.Idle || bat.State.Raw == battery.Unknown {
			pluggedIn = true
		}
	}
	return !pluggedIn && discharging
}

func averageBatteryLevel(batteries []battery.Battery) float64 {
	var total, count float64
	for _, bat := range batteries {
		if bat.Full <= 0 {
			continue
		}
		count++
		total += bat.Current * 100 / bat.Full
	}
	if count == 0 {
		return 0
	}
	return total / count
}
