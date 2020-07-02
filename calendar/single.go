package calendar

type SingleValue struct {
	singleValue    uint
	hasSingleValue bool
}

func (v *SingleValue) HasSingleValue() bool {
	return v.hasSingleValue
}
