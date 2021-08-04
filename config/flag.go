package config

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	emptyStringArray []string
)

func init() {
	emptyStringArray = make([]string, 0)
}

func convertStructToFlags(orig interface{}) map[string][]string {
	typeOf := reflect.TypeOf(orig)
	valueOf := reflect.ValueOf(orig)

	if typeOf.Kind() == reflect.Ptr {
		// Deference the pointer
		typeOf = typeOf.Elem()
		valueOf = valueOf.Elem()
	}

	flags := make(map[string][]string)
	// NumField() will panic if typeOf is not a struct
	if typeOf.Kind() != reflect.Struct {
		panic(fmt.Errorf("unsupported type %s, expected %s", typeOf.Kind(), reflect.Struct))
	}
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if argument, ok := field.Tag.Lookup("argument"); ok {
			if argument != "" {
				convert, ok := stringifyConfidentialValue(valueOf.Field(i))
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

// stringifyConfidentialValue returns a string representation of the value including confidential parts
func stringifyConfidentialValue(value reflect.Value) ([]string, bool) {
	if value.Type() == reflect.TypeOf(ConfidentialValue{}) {
		method := value.MethodByName("Value")
		if method.IsValid() {
			values := method.Call([]reflect.Value{})
			if len(values) == 1 {
				value = values[0]
			}
		}
	}
	return stringifyValue(value)
}

// stringifyValue returns a string representation of the value, and if it has any value at all
func stringifyValue(value reflect.Value) ([]string, bool) {
	// Check if the value can convert itself to String() (e.g. time.Duration)
	stringer := fmt.Stringer(nil)
	if value.CanInterface() {
		vi := value.Interface()
		if s, ok := vi.(fmt.Stringer); ok {
			stringer = s
		}
	}

	var stringVal string

	switch value.Kind() {
	case reflect.String:
		if stringer != nil {
			stringVal = stringer.String()
		} else {
			stringVal = value.String()
		}
		return []string{stringVal}, stringVal != ""

	case reflect.Bool:
		boolVal := value.Bool()
		return emptyStringArray, boolVal

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := value.Int()
		if stringer != nil {
			stringVal = stringer.String()
		} else {
			stringVal = strconv.FormatInt(intVal, 10)
		}
		return []string{stringVal}, intVal != 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intVal := value.Uint()
		if stringer != nil {
			stringVal = stringer.String()
		} else {
			stringVal = strconv.FormatUint(intVal, 10)
		}
		return []string{stringVal}, intVal != 0

	case reflect.Float32, reflect.Float64:
		floatVal := value.Float()
		if stringer != nil {
			stringVal = stringer.String()
		} else {
			stringVal = strconv.FormatFloat(floatVal, 'f', -1, 64)
		}
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
				panic(fmt.Errorf("array of array of values are not supported"))
			}
			sliceVal[i] = v[0]
		}
		return sliceVal, true

	case reflect.Interface:
		return stringifyValue(value.Elem())

	default:
		if stringer != nil {
			stringVal = stringer.String()
			return []string{stringVal}, stringVal != ""
		}
		return []string{fmt.Sprintf("ERROR: unexpected type %s", value.Kind())}, false
	}
}
