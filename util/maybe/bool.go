package maybe

type Bool struct {
	Optional[bool]
}

func False() Bool {
	return Bool{Set(false)}
}

func True() Bool {
	return Bool{Set(true)}
}

func (value Bool) IsTrue() bool {
	return value.HasValue() && value.Value()
}

func (value Bool) IsStrictlyFalse() bool {
	return value.HasValue() && value.Value() == false
}

func (value Bool) IsFalseOrUndefined() bool {
	return !value.HasValue() || value.Value() == false
}

func (value Bool) IsUndefined() bool {
	return !value.HasValue()
}

func (value Bool) IsTrueOrUndefined() bool {
	return !value.HasValue() || value.Value()
}
