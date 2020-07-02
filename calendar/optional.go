package calendar

type OptionalValue struct {
	hasValue bool
}

func (v *OptionalValue) HasValue() bool {
	return v.hasValue
}
