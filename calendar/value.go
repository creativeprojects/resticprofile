package calendar

type Value struct {
	*OptionalValue
	*SingleValue
	*Range
	singleValue uint
}

func NewValue(min, max uint) *Value {
	return &Value{
		OptionalValue: &OptionalValue{},
		SingleValue:   &SingleValue{},
		Range:         NewRange(min, max),
	}
}

func (v *Value) AddValue(value uint) {
	if !v.hasValue {
		// 1st time: no value here before
		v.addSingleValue(value)
		return
	}
	if v.hasSingleValue {
		// 2nd time: single value here before
		v.hasSingleValue = false
		v.addRangeValue(v.singleValue)
		v.singleValue = 0
	}
	v.addRangeValue(value)
}

func (r *Value) AddRange(min uint, max uint) {
	for i := min; i <= max; i++ {
		r.AddValue(i)
	}
}

func (v *Value) addSingleValue(value uint) {
	v.hasValue = true
	v.hasSingleValue = true
	v.singleValue = value
}
