package calendar

import "fmt"

type Value struct {
	hasValue       bool
	hasSingleValue bool
	hasRange       bool
	singleValue    int
	rangeValues    []bool
	minRange       int
	maxRange       int
}

func NewValue(min, max int) *Value {
	return &Value{
		minRange: min,
		maxRange: max,
	}
}

func (v *Value) HasValue() bool {
	return v.hasValue
}

func (v *Value) HasSingleValue() bool {
	return v.hasSingleValue
}

func (v *Value) HasRange() bool {
	return v.hasRange
}

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

func (v *Value) AddValue(value int) {
	if !v.hasValue {
		// 1st time: no value here before
		v.addSingleValue(value)
		return
	}
	if v.hasSingleValue {
		// 2nd time: single value here before
		v.addRangeValue(v.singleValue)
		v.hasSingleValue = false
		v.singleValue = 0
	}
	v.addRangeValue(value)
}

func (v *Value) AddRange(min int, max int) {
	for i := min; i <= max; i++ {
		v.AddValue(i)
	}
}

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

func (v *Value) GetRanges() []struct{ start, end int } {
	if !v.hasValue {
		return []struct{ start, end int }{}
	}

	if v.hasSingleValue {
		return []struct{ start, end int }{
			{
				start: v.singleValue,
				end:   v.singleValue,
			},
		}
	}

	ranges := make([]struct{ start, end int }, 0, 1)
	return ranges
}

func (v *Value) initRange() {
	v.rangeValues = make([]bool, v.maxRange-v.minRange+1)
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
	if value < v.minRange {
		panic(fmt.Sprintf("Value outside of range: %d is lower than %d", value, v.minRange))
	}
	if value > v.maxRange {
		panic(fmt.Sprintf("Value outside of range: %d is greater than %d", value, v.maxRange))
	}
	v.rangeValues[value-v.minRange] = true
	v.hasRange = true
}
