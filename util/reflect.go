package util

import "reflect"

// UnpackValue dereferences reflect.Pointer and reflect.Interface values
func UnpackValue(value reflect.Value) (v reflect.Value, isNil bool) {
	v = value
	for v.Kind() == reflect.Pointer ||
		v.Kind() == reflect.Interface {
		isNil = isNil || v.IsNil()
		v = v.Elem()
	}
	return
}

// ElementType returns the type after removing all reflect.Array, reflect.Chan, reflect.Map, reflect.Pointer and reflect.Slice
func ElementType(initial reflect.Type) (et reflect.Type) {
	et = initial
	for et.Kind() == reflect.Array ||
		et.Kind() == reflect.Chan ||
		et.Kind() == reflect.Map ||
		et.Kind() == reflect.Pointer ||
		et.Kind() == reflect.Slice {
		et = et.Elem()
	}
	return
}
