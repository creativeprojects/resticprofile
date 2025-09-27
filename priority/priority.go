package priority

// Priority class.
// Can be used as a priority class in Windows, or as a "nice" value in unixes
const (
	Idle       = 19
	Background = 15
	Low        = 10
	Normal     = 0
	High       = -10
	Highest    = -20
)

var (
	// Values is the map between the name and the value
	Values = map[string]int{
		"idle":       Idle,
		"background": Background,
		"low":        Low,
		"normal":     Normal,
		"high":       High,
		"highest":    Highest,
	}
)
