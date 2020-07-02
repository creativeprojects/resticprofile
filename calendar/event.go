package calendar

// Should be able to read the same calendar events
// https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events

// Event represents a calendar event.
// It can be one specific point in time, or a recurring event
type Event struct {
	WeekDay [7]bool
	Year    *Value
	Month   *Value
	Day     *Value
	Hour    *Value
	Minute  *Value
	Second  *Value
}

func NewEvent() *Event {
	return &Event{
		Year:   NewValue(2000, 2200),
		Month:  NewValue(1, 12),
		Day:    NewValue(1, 31),
		Hour:   NewValue(0, 23),
		Minute: NewValue(0, 59),
		Second: NewValue(0, 59),
	}
}
