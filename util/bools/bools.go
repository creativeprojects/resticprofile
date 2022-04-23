package bools

import "github.com/creativeprojects/resticprofile/util"

func IsTrue(value *bool) bool {
	return util.NotNilAnd(value, true)
}

func IsStrictlyFalse(value *bool) bool {
	return util.NotNilAnd(value, false)
}

func IsFalseOrUndefined(value *bool) bool {
	return util.NilOr(value, false)
}

func IsUndefined(value *bool) bool {
	return value == nil
}

func IsTrueOrUndefined(value *bool) bool {
	return util.NilOr(value, true)
}

func False() *bool {
	return util.CopyRef(false)
}

func True() *bool {
	return util.CopyRef(true)
}
