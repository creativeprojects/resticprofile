//go:build darwin
// +build darwin

package schedule

type HandlerLaunchd struct {
	//
}

func NewHandler() *HandlerLaunchd {
	return &HandlerLaunchd{}
}

// Init verifies launchd is available on this system
func (h *HandlerLaunchd) Init() error {
	return lookupBinary("launchd", launchdBin)
}

// Close does nothing with launchd
func (h *HandlerLaunchd) Close() {}

var (
	_ Handler = &HandlerLaunchd{}
)

// CalendarInterval contains date and time trigger definition inside a map.
// keys of the map should be:
//  "Month"   Month of year (1..12, 1 being January)
// 	"Day"     Day of month (1..31)
// 	"Weekday" Day of week (0..7, 0 and 7 being Sunday)
// 	"Hour"    Hour of day (0..23)
// 	"Minute"  Minute of hour (0..59)
type CalendarInterval map[string]int

// newCalendarInterval creates a new map of 5 elements
func newCalendarInterval() *CalendarInterval {
	var value CalendarInterval = make(map[string]int, 5)
	return &value
}

func (c *CalendarInterval) clone() *CalendarInterval {
	clone := newCalendarInterval()
	for key, value := range *c {
		(*clone)[key] = value
	}
	return clone
}
