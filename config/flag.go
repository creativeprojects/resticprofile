package config

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/shell"
)

var (
	emptyStringArray []string
)

func init() {
	emptyStringArray = make([]string, 0)
}

func convertStructToArgs(orig interface{}, args *shell.Args) *shell.Args {
	typeOf := reflect.TypeOf(orig)
	valueOf := reflect.ValueOf(orig)

	if typeOf.Kind() == reflect.Ptr {
		// Deference the pointer
		typeOf = typeOf.Elem()
		valueOf = valueOf.Elem()
	}

	if args == nil {
		args = shell.NewArgs()
	}
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
					argType := shell.ArgConfigEscape
					// check if the argument type was specified
					rawType, ok := field.Tag.Lookup("argument-type")
					if ok && rawType == "no-glob" {
						argType = shell.ArgConfigKeepGlobQuote
					}
					args.AddFlags(argument, convert, argType)
				}
			}
		}
	}
	return args
}

func addOtherArgs(args *shell.Args, otherArgs map[string]interface{}) *shell.Args {
	if len(otherArgs) == 0 {
		return args
	}

	// Add other args
	for name, value := range otherArgs {
		if name == constants.SectionConfigurationMixinUse {
			continue
		}
		if convert, ok := stringifyValueOf(value); ok {
			args.AddFlags(name, convert, shell.ArgConfigEscape)
		}
	}
	return args
}

// stringifyValueOf returns a string representation of the value, and if it has any value at all
func stringifyValueOf(value interface{}) ([]string, bool) {
	return stringifyAnyValueOf(value, true)
}

// stringifyAnyValueOf returns a string representation of the value, and if it has any value at all
func stringifyAnyValueOf(value interface{}, onlySimplyValues bool) ([]string, bool) {
	if value == nil {
		return emptyStringArray, false
	}
	return stringifyAnyValue(reflect.ValueOf(value), onlySimplyValues)
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
	return stringifyAnyValue(value, true)
}

// stringifyAnyValue returns a string representation of the value, and if it has any value at all
func stringifyAnyValue(value reflect.Value, onlySimplyValues bool) ([]string, bool) {
	// Check if the value can convert itself to String() (e.g. time.Duration)
	stringer := fmt.Stringer(nil)
	if value.Kind() != reflect.Invalid && value.CanInterface() {
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
		for i := 0; i < n; i++ {
			v, _ := stringifyAnyValue(value.Index(i), onlySimplyValues)
			if len(v) > 1 {
				if onlySimplyValues {
					panic(fmt.Errorf("array of array of values are not supported"))
				}
				sliceVal[i] = "{" + strings.Join(v, ",") + "}"
			} else {
				sliceVal[i] = v[0]
			}
		}
		return sliceVal, n > 0

	case reflect.Map:
		if onlySimplyValues {
			return []string{fmt.Sprintf("ERROR: unexpected type %s", reflect.Map)}, false
		}
		flatMap := make([]string, 0, value.Len())
		for it := value.MapRange(); it.Next(); {
			k, _ := stringifyAnyValue(it.Key(), false)
			v, hasValue := stringifyAnyValue(it.Value(), false)
			if len(v) > 1 {
				v[0] = "{" + strings.Join(v, ",") + "}"
			}
			if len(v) == 0 && hasValue {
				v = []string{"true"}
			}
			if len(k) == 1 && len(v) > 0 {
				flatMap = append(flatMap, fmt.Sprintf("%s:%s", k[0], v[0]))
			}
		}
		sort.Strings(flatMap)
		return flatMap, len(flatMap) > 0

	case reflect.Interface:
		return stringifyAnyValue(value.Elem(), onlySimplyValues)

	default:
		if stringer != nil {
			stringVal = stringer.String()
			return []string{stringVal}, stringVal != ""
		}
		return []string{fmt.Sprintf("ERROR: unexpected type %s", value.Kind())}, false
	}
}
