package batt

import (
	"github.com/distatus/battery"
)

// Batteries return all connected batteries information
func Batteries() ([]battery.Battery, error) {
	return nil, ErrBatteryInfoNotSupported
}

// IsRunningOnBattery returns true if the computer is running on battery,
// followed by the percentage of battery remaining. If no battery is present
// the percentage will be 0.
func IsRunningOnBattery() (bool, int, error) {
	return false, 0, ErrBatteryInfoNotSupported
}
