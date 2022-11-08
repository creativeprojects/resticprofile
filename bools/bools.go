package bools

func IsTrue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func IsStrictlyFalse(value *bool) bool {
	if value == nil {
		return false
	}
	return !*value
}

func IsFalseOrUndefined(value *bool) bool {
	if value == nil {
		return true
	}
	return !*value
}

func IsUndefined(value *bool) bool {
	return value == nil
}

func IsTrueOrUndefined(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}
