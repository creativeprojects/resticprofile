package calendar

const (
	minDay = 0
	maxDay = 8
)

var (
	shortWeekDay = [maxDay]string{
		"sun",
		"mon",
		"tue",
		"wed",
		"thu",
		"fri",
		"sat",
		"sun",
	}

	longWeekDay = [maxDay]string{
		"sunday",
		"monday",
		"tuesday",
		"wednesday",
		"thusday",
		"friday",
		"saturday",
		"sunday",
	}

	specialKeywords = map[string]func(event *Event){
		"minutely": func(event *Event) {
			event.Second.MustAddValue(0)
		},
		"hourly": func(event *Event) {
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"daily": func(event *Event) {
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"weekly": func(event *Event) {
			event.WeekDay.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"monthly": func(event *Event) {
			event.Day.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"quarterly": func(event *Event) {
			event.Month.MustAddValue(1)
			event.Month.MustAddValue(4)
			event.Month.MustAddValue(7)
			event.Month.MustAddValue(10)
			event.Day.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"semiannually": func(event *Event) {
			event.Month.MustAddValue(1)
			event.Month.MustAddValue(7)
			event.Day.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"yearly": func(event *Event) {
			event.Month.MustAddValue(1)
			event.Day.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
		"annually": func(event *Event) {
			event.Month.MustAddValue(1)
			event.Day.MustAddValue(1)
			event.Hour.MustAddValue(0)
			event.Minute.MustAddValue(0)
			event.Second.MustAddValue(0)
		},
	}
)
