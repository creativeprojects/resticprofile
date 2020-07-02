package calendar

var (
	shortWeekDay = [7]string{
		"mon",
		"tue",
		"wed",
		"thu",
		"fri",
		"sat",
		"sun",
	}

	longWeekDay = [7]string{
		"monday",
		"tuesday",
		"wednesday",
		"thusday",
		"friday",
		"saturday",
		"sunday",
	}

	specialKeywords = map[string]*Event{
		"minutely": NewEvent(func(event *Event) {
			event.Second.AddValue(0)
		}),
		"hourly": NewEvent(func(event *Event) {
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"daily": NewEvent(func(event *Event) {
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"weekly": NewEvent(func(event *Event) {
			event.WeekDay.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"monthly": NewEvent(func(event *Event) {
			event.Day.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"quarterly": NewEvent(func(event *Event) {
			event.Month.AddValue(1)
			event.Month.AddValue(4)
			event.Month.AddValue(7)
			event.Month.AddValue(10)
			event.Day.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"semiannually": NewEvent(func(event *Event) {
			event.Month.AddValue(1)
			event.Month.AddValue(7)
			event.Day.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"yearly": NewEvent(func(event *Event) {
			event.Month.AddValue(1)
			event.Day.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
		"annually": NewEvent(func(event *Event) {
			event.Month.AddValue(1)
			event.Day.AddValue(1)
			event.Hour.AddValue(0)
			event.Minute.AddValue(0)
			event.Second.AddValue(0)
		}),
	}
)
