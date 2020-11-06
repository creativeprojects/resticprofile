package calendar

import (
	"fmt"
	"strconv"
	"strings"
)

type postProcessFunc func(int) (int, error)

// TypeValue represents the type of a Value
type TypeValue int

// TypeValue
const (
	TypeUnknown TypeValue = iota
	TypeWeekDay
	TypeYear
	TypeMonth
	TypeDay
	TypeHour
	TypeMinute
	TypeSecond
)

// Range represents a range of values: from Range.Start to Range.End
type Range struct {
	Start int
	End   int
}

// Value is represented by either no value, a single value, or a range of values
type Value struct {
	definedType    TypeValue
	hasValue       bool
	hasSingleValue bool
	hasRange       bool
	singleValue    int
	rangeValues    []bool
	minRange       int
	maxRange       int
}

// NewValue creates a new value
func NewValue(min, max int) *Value {
	return &Value{
		minRange: min,
		maxRange: max,
	}
}

// NewValueFromType creates a new value from a predefined type
func NewValueFromType(t TypeValue) *Value {
	min, max := 0, 0
	switch t {
	case TypeWeekDay:
		min = minDay
		max = maxDay - 1
	case TypeYear:
		min = 2000
		max = 2200
	case TypeMonth:
		min = 1
		max = 12
	case TypeDay:
		min = 1
		max = 31
	case TypeHour:
		max = 23
	case TypeMinute:
		max = 59
	case TypeSecond:
		max = 59
	}
	return &Value{
		definedType: t,
		minRange:    min,
		maxRange:    max,
	}
}

// HasValue is true when one or more value has been set
func (v *Value) HasValue() bool {
	return v.hasValue
}

// HasSingleValue is true when a exactly one value has been set
func (v *Value) HasSingleValue() bool {
	return v.hasSingleValue
}

// HasRange is true when at least two values has been set
func (v *Value) HasRange() bool {
	return v.hasRange
}

// HasContiguousRange is true when two or more values are contiguous
func (v *Value) HasContiguousRange() bool {
	if !v.hasRange {
		return false
	}

	for i := 0; i < v.maxRange-v.minRange; i++ {
		if v.rangeValues[i] && v.rangeValues[i+1] {
			return true
		}
	}
	return false
}

// HasLongContiguousRange is true when three or more values are contiguous
func (v *Value) HasLongContiguousRange() bool {
	if !v.hasRange {
		return false
	}

	for i := 0; i < v.maxRange-v.minRange-1; i++ {
		if v.rangeValues[i] && v.rangeValues[i+1] && v.rangeValues[i+2] {
			return true
		}
	}
	return false
}

// GetType returns the defined type
func (v *Value) GetType() TypeValue {
	return v.definedType
}

// MustAddValue adds a new value and panics if an error arises
func (v *Value) MustAddValue(value int) {
	err := v.AddValue(value)
	if err != nil {
		panic(err)
	}
}

// AddValue adds a new value
func (v *Value) AddValue(value int) error {
	err := v.checkValue(value)
	if err != nil {
		return err
	}
	if !v.hasValue {
		// 1st time: no value here before
		v.addSingleValue(value)
		return nil
	}
	if v.hasSingleValue {
		// 2nd time: single value here before
		v.addRangeValue(v.singleValue)
		v.hasSingleValue = false
		v.singleValue = 0
	}
	v.addRangeValue(value)
	return nil
}

// MustAddRange adds a range of values from min to max and panics if an error occurs
func (v *Value) MustAddRange(min int, max int) {
	err := v.AddRange(min, max)
	if err != nil {
		panic(err)
	}
}

// AddRange adds a range of values from start to end
func (v *Value) AddRange(start int, end int) error {
	if end < start {
		return v.AddRange(end, start)
	}
	for i := start; i <= end; i++ {
		err := v.AddValue(i)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRangeValues returns a list of values
func (v *Value) GetRangeValues() []int {
	if !v.hasValue {
		return []int{}
	}

	if v.hasSingleValue {
		return []int{v.singleValue}
	}

	values := []int{}
	for i := 0; i <= v.maxRange-v.minRange; i++ {
		if v.rangeValues[i] {
			values = append(values, i+v.minRange)
		}
	}
	return values
}

// GetRanges returns a list or range of values (from Range.Start to Range.End)
func (v *Value) GetRanges() []Range {
	if !v.hasValue {
		return []Range{}
	}

	if v.hasSingleValue {
		return []Range{
			{
				Start: v.singleValue,
				End:   v.singleValue,
			},
		}
	}

	var currentRange *Range
	ranges := make([]Range, 0, 1)
	for i := 0; i <= v.maxRange-v.minRange; i++ {
		if v.rangeValues[i] {
			// there's a value
			if currentRange == nil {
				// new range
				currentRange = &Range{
					Start: i + v.minRange,
					End:   i + v.minRange,
				}
				continue
			}
			// extend this range
			currentRange.End = i + v.minRange
			continue
		}
		// no value here
		if currentRange != nil {
			// that was the end of this one
			ranges = append(ranges, *currentRange)
			currentRange = nil
		}
	}
	// a last one at the end?
	if currentRange != nil {
		// that was the end of this one
		ranges = append(ranges, *currentRange)
	}
	return ranges
}

// IsInRange check the parameter is in range of Value
func (v *Value) IsInRange(ref int) bool {
	for _, current := range v.GetRangeValues() {
		if current == ref {
			return true
		}
	}
	return false
}

// String representation
func (v *Value) String() string {
	if !v.hasValue {
		return "*"
	}
	if v.hasSingleValue {
		return fmt.Sprintf("%02d", v.singleValue)
	}
	output := []string{}
	if !v.HasLongContiguousRange() {
		for _, r := range v.GetRangeValues() {
			output = append(output, fmt.Sprintf("%02d", r))
		}
	} else {
		for _, r := range v.GetRanges() {
			if r.Start == r.End {
				output = append(output, fmt.Sprintf("%02d", r.Start))
				continue
			}
			output = append(output, fmt.Sprintf("%02d..%02d", r.Start, r.End))
		}
	}
	return strings.Join(output, ",")
}

// Parse a string into a value
func (v *Value) Parse(input string, postProcess ...postProcessFunc) error {
	// clear up data first
	v.hasValue = false
	v.hasSingleValue = false
	v.hasRange = false

	if input == "*" {
		// done!
		return nil
	}

	parts := strings.Split(input, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}
		err := v.parseUnit(part, postProcess...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Value) parseUnit(input string, postProcess ...postProcessFunc) error {
	if strings.Contains(input, "..") {
		// this is a range
		var start, end int
		parsed, err := fmt.Sscanf(input, "%d..%d", &start, &end)
		if err != nil {
			return err
		}
		if parsed != 2 {
			return fmt.Errorf("cannot parse range '%s'", input)
		}
		// run post-processing functions before adding the value
		start, err = runPostProcess(start, postProcess)
		if err != nil {
			return err
		}
		end, err = runPostProcess(end, postProcess)
		if err != nil {
			return err
		}
		// now push the value
		err = v.AddRange(start, end)
		if err != nil {
			return err
		}
		return nil
	}
	i, err := parseInt(input)
	if err != nil {
		return err
	}
	// run post-processing functions before adding the value
	i, err = runPostProcess(i, postProcess)
	if err != nil {
		return err
	}
	// now push the value
	err = v.AddValue(i)
	if err != nil {
		return err
	}
	return nil
}

func parseInt(input string) (int, error) {
	i, err := strconv.ParseInt(input, 10, 32)
	return int(i), err
}

func runPostProcess(value int, postProcess []postProcessFunc) (int, error) {
	if len(postProcess) == 0 {
		return value, nil
	}
	var err error
	for _, f := range postProcess {
		value, err = f(value)
		if err != nil {
			return value, err
		}
	}
	return value, nil
}

func (v *Value) initRange() {
	v.rangeValues = make([]bool, v.maxRange-v.minRange+1)
}

func (v *Value) checkValue(value int) error {
	if value < v.minRange {
		return fmt.Errorf("value outside of range: %d is lower than %d", value, v.minRange)
	}
	if value > v.maxRange {
		return fmt.Errorf("value outside of range: %d is greater than %d", value, v.maxRange)
	}
	return nil
}

func (v *Value) addSingleValue(value int) {
	v.hasValue = true
	v.hasSingleValue = true
	v.singleValue = value
}

func (v *Value) addRangeValue(value int) {
	if !v.hasRange {
		// first time here, we initialize the slice
		v.initRange()
	}
	v.rangeValues[value-v.minRange] = true
	v.hasRange = true
}
