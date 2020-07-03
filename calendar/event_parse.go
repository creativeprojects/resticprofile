package calendar

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	unit             = "[0-9*.,]+"
	weekday          = "([a-zA-Z0-9*.,]+)"
	datePattern      = "(" + unit + "-|)(" + unit + ")-(" + unit + ")" // year or nothing then month then day
	timePattern      = "(" + unit + "):(" + unit + ")(:" + unit + "|)" // hour, minute then second or nothing
	yearMonthDay     = "(" + unit + ")-(" + unit + ")-(" + unit + ")"
	monthDay         = "(" + unit + ")-(" + unit + ")"
	hourMinuteSecond = "(" + unit + "):(" + unit + "):(" + unit + ")"
	hourMinute       = "(" + unit + "):(" + unit + ")"
)

type parseFunc func(e *Event, match []string) error

var (
	regexpWeekdayFullDateTime = regexp.MustCompile("^" + weekday + " " + yearMonthDay + " " + hourMinuteSecond + "$")
	regexpWeekdayFullTime     = regexp.MustCompile("^" + weekday + " " + hourMinuteSecond + "$")
	regexpFullDateTime        = regexp.MustCompile("^" + yearMonthDay + " " + hourMinuteSecond + "$")
	regexpFullDateHourMinute  = regexp.MustCompile("^" + yearMonthDay + " " + hourMinute + "$")
	regexpYearMonthDay        = regexp.MustCompile("^" + yearMonthDay + "$")
	regexpMonthDay            = regexp.MustCompile("^" + monthDay + "$")

	regexpWeekdayPattern     = regexp.MustCompile("^" + weekday + "$")
	regexpDatePattern        = regexp.MustCompile("^" + datePattern + "$")
	regexpTimePattern        = regexp.MustCompile("^" + timePattern + "$")
	regexpWeekdayDatePattern = regexp.MustCompile("^" + weekday + " " + datePattern + "$")
	regexpWeekdayTimePattern = regexp.MustCompile("^" + weekday + " " + timePattern + "$")
	regexpDateTimePattern    = regexp.MustCompile("^" + datePattern + " " + timePattern + "$")
	regexpFullPattern        = regexp.MustCompile("^" + weekday + " " + datePattern + " " + timePattern + "$")

	// parsingRules are the rules for parsing each field from regular expression match
	parsingRules = []struct {
		expr        *regexp.Regexp
		parseValues []parseFunc
	}{
		// {regexpWeekdayFullDateTime, []parseFunc{parseWeekday(1), parseYear(2), parseMonth(3), parseDay(4), parseHour(5), parseMinute(6), parseSecond(7)}},
		// {regexpWeekdayFullTime, []parseFunc{parseWeekday(1), parseHour(2), parseMinute(3), parseSecond(4)}},
		// {regexpFullDateTime, []parseFunc{parseYear(1), parseMonth(2), parseDay(3), parseHour(4), parseMinute(5), parseSecond(6)}},
		// {regexpFullDateHourMinute, []parseFunc{parseYear(1), parseMonth(2), parseDay(3), parseHour(4), parseMinute(5), setZeroSecond()}},
		// {regexpYearMonthDay, []parseFunc{parseYear(1), parseMonth(2), parseDay(3), setMidnight()}},
		// {regexpMonthDay, []parseFunc{parseMonth(1), parseDay(2), setMidnight()}},
		{regexpFullPattern, []parseFunc{parseWeekday(1), parseYear(2), parseMonth(3), parseDay(4), parseHour(5), parseMinute(6), parseSecond(7)}},
		{regexpDatePattern, []parseFunc{parseYear(1), parseMonth(2), parseDay(3), setMidnight()}},
		{regexpTimePattern, []parseFunc{parseHour(1), parseMinute(2), parseSecond(3)}},
		{regexpDateTimePattern, []parseFunc{parseYear(1), parseMonth(2), parseDay(3), parseHour(4), parseMinute(5), parseSecond(6)}},
		{regexpWeekdayPattern, []parseFunc{parseWeekday(1), setMidnight()}},
		{regexpWeekdayDatePattern, []parseFunc{parseWeekday(1), parseYear(2), parseMonth(3), parseDay(4), setMidnight()}},
		{regexpWeekdayTimePattern, []parseFunc{parseWeekday(1), parseHour(2), parseMinute(3), parseSecond(4)}},
	}
)

func parseYear(index int) parseFunc {
	return func(e *Event, match []string) error {
		// year can be empty => it means any year
		if match[index] == "" {
			return nil
		}
		return e.Year.Parse(strings.Trim(match[index], "-"))
	}
}

func parseMonth(index int) parseFunc {
	return func(e *Event, match []string) error {
		return e.Month.Parse(match[index])
	}
}

func parseDay(index int) parseFunc {
	return func(e *Event, match []string) error {
		return e.Day.Parse(match[index])
	}
}

func parseHour(index int) parseFunc {
	return func(e *Event, match []string) error {
		return e.Hour.Parse(match[index])
	}
}

func parseMinute(index int) parseFunc {
	return func(e *Event, match []string) error {
		return e.Minute.Parse(match[index])
	}
}

func parseSecond(index int) parseFunc {
	return func(e *Event, match []string) error {
		// second can be empty => it means zero
		if match[index] == "" {
			e.Second.AddValue(0)
			return nil
		}
		return e.Second.Parse(strings.Trim(match[index], ":"))
	}
}

func setMidnight() parseFunc {
	return func(e *Event, match []string) error {
		e.Hour.AddValue(0)
		e.Minute.AddValue(0)
		e.Second.AddValue(0)
		return nil
	}
}

func setZeroSecond() parseFunc {
	return func(e *Event, match []string) error {
		e.Second.AddValue(0)
		return nil
	}
}

func parseWeekday(index int) parseFunc {
	return func(e *Event, match []string) error {
		weekdays := strings.ToLower(match[index])
		for dayIndex, day := range longWeekDay {
			weekdays = strings.ReplaceAll(weekdays, day, fmt.Sprintf("%02d", dayIndex+1))
		}
		for dayIndex, day := range shortWeekDay {
			weekdays = strings.ReplaceAll(weekdays, day, fmt.Sprintf("%02d", dayIndex+1))
		}
		return e.WeekDay.Parse(weekdays)
	}
}
