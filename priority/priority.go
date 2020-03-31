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
