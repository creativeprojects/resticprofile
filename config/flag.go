package config

import (
	"fmt"
	"reflect"
)

var (
	emptyStringArray []string
)

func init() {
	emptyStringArray = make([]string, 0)
}

func convertStructToFlags(orig interface{}) map[string][]string {
	flags := make(map[string][]string, 0)
	pt := reflect.TypeOf(orig)

	// NumField() will panic if pt is not a struct
	if pt.Kind() != reflect.Struct {
		return flags
	}
	for i := 0; i < pt.NumField(); i++ {
		field := pt.Field(i)
		if argument, ok := field.Tag.Lookup("argument"); ok {
			if argument != "" {
				convert, ok := stringifyValueOf("set") // <-- find value of field
				if ok {
					flags[argument] = convert
				}
			}
		}
	}
	return flags
}

// stringifyValueOf returns a string representation of the value, and if it has any value at all
func stringifyValueOf(value interface{}) ([]string, bool) {
	if value == nil {
		return emptyStringArray, false
	}
	return stringifyValue(reflect.ValueOf(value))
}

// stringifyValue returns a string representation of the value, and if it has any value at all
func stringifyValue(value reflect.Value) ([]string, bool) {

	switch value.Kind() {
	case reflect.String:
		stringVal := value.String()
		return []string{stringVal}, stringVal != ""

	case reflect.Bool:
		boolVal := value.Bool()
		return emptyStringArray, boolVal

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := value.Int()
		stringVal := fmt.Sprintf("%d", intVal)
		return []string{stringVal}, intVal != 0

	case reflect.Float32, reflect.Float64:
		floatVal := value.Float()
		stringVal := fmt.Sprintf("%f", floatVal)
		return []string{stringVal}, floatVal != 0

	case reflect.Slice, reflect.Array:
		n := value.Len()
		sliceVal := make([]string, n)
		if n == 0 {
			return sliceVal, false
		}
		for i := 0; i < n; i++ {
			v, _ := stringifyValue(value.Index(i))
			if len(v) > 1 {
				panic(fmt.Errorf("Array of array of values are not supported"))
			}
			sliceVal[i] = v[0]
		}
		return sliceVal, true

	case reflect.Interface:
		return stringifyValue(value.Elem())

	default:
		panic(fmt.Errorf("Unexpected type %s", value.Kind()))
	}
}
