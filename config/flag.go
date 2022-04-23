package config

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/util"
)

var (
	emptyStringArray []string
)

func init() {
	emptyStringArray = make([]string, 0)
}

func addArgsFromStruct(args *shell.Args, section any) {
	valueOf, isNil := util.UnpackValue(reflect.ValueOf(section))
	if isNil {
		return
	} else if valueOf.Kind() != reflect.Struct {
		panic(fmt.Errorf("unsupported type %s, expected %s", valueOf.Kind(), reflect.Struct))
	}

	typeOf := valueOf.Type()
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
}

func argAliasesFromStruct(section any) (aliases map[string]string) {
	aliases = make(map[string]string)
	if t := util.ElementType(reflect.TypeOf(section)); t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if argument, ok := field.Tag.Lookup("argument"); ok {
				if alias, ok := field.Tag.Lookup("mapstructure"); ok && alias != argument {
					aliases[alias] = argument
				}
			}
		}
	}
	return aliases
}

func addArgsFromMap(args *shell.Args, argAliases map[string]string, argsMap map[string]any) {
	// Add other args
	for name, value := range argsMap {
		if name == constants.SectionConfigurationMixinUse {
			continue
		}
		if convert, ok := stringifyValueOf(value); ok {
			if targetName, found := argAliases[name]; found {
				name = targetName
			}
			args.AddFlags(name, convert, shell.ArgConfigEscape)
		}
	}
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
	return stringify(reflect.ValueOf(value), onlySimplyValues)
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
	return stringify(value, true)
}

// Returns the stringer of the given value or nil if the value doesn't implement fmt.Stringer
func getStringer(value reflect.Value) (stringer fmt.Stringer) {
	if value.Kind() != reflect.Invalid && value.CanInterface() {
		vi := value.Interface()
		if s, ok := vi.(fmt.Stringer); ok {
			stringer = s
		}
	}
	return
}

// stringify returns a string representation of the value, and if it has any value at all
func stringify(value reflect.Value, onlySimplyValues bool) ([]string, bool) {
	// Check if the value can convert itself to String() (e.g. time.Duration)
	stringer := getStringer(value)

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
			v, _ := stringify(value.Index(i), onlySimplyValues)
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
			k, _ := stringify(it.Key(), false)
			v, hasValue := stringify(it.Value(), false)
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
		return stringify(value.Elem(), onlySimplyValues)

	default:
		if stringer != nil {
			stringVal = stringer.String()
			return []string{stringVal}, stringVal != ""
		}
		return []string{fmt.Sprintf("ERROR: unexpected type %s", value.Kind())}, false
	}
}
