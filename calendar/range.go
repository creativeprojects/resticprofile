package calendar

import "fmt"

type Range struct {
	rangeValues []bool
	minRange    uint
	maxRange    uint
	hasRange    bool
}

func NewRange(min, max uint) *Range {
	return &Range{
		minRange: min,
		maxRange: max,
	}
}

func (r *Range) initRange() {
	r.rangeValues = make([]bool, r.maxRange-r.minRange+1)
}

func (r *Range) HasRange() bool {
	return r.hasRange
}

func (r *Range) getRangeValues() []uint {
	values := []uint{}
	for i := uint(0); i <= r.maxRange-r.minRange; i++ {
		if r.rangeValues[i] {
			values = append(values, i+r.minRange)
		}
	}
	return values
}

func (r *Range) addRangeValue(value uint) {
	if !r.hasRange {
		// first time here, we initialize the slice
		r.initRange()
	}
	if value < r.minRange {
		panic(fmt.Sprintf("Value outside of range: %d is lower than %d", value, r.minRange))
	}
	if value > r.maxRange {
		panic(fmt.Sprintf("Value outside of range: %d is greater than %d", value, r.maxRange))
	}
	r.rangeValues[value-r.minRange] = true
	r.hasRange = true
}
