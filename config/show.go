package config

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
)

// ShowStruct write out to w a human readable text representation of the orig parameter
func ShowStruct(w io.Writer, orig interface{}, name string) error {
	display := newDisplay(name, w)
	err := showSubStruct(orig, []string{}, display)
	if err != nil {
		return err
	}
	display.Flush()
	return nil
}

func showSubStruct(orig interface{}, stack []string, display *Display) error {
	typeOf := reflect.TypeOf(orig)
	valueOf := reflect.ValueOf(orig)

	if typeOf.Kind() == reflect.Ptr {
		// Deference the pointer
		typeOf = typeOf.Elem()
		valueOf = valueOf.Elem()
	}

	// NumField() will panic if typeOf is not a struct
	if typeOf.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type %s, expected %s", typeOf.Kind(), reflect.Struct)
	}

	for i := 0; i < typeOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		fieldValue := valueOf.Field(i)

		err := showField(fieldType, fieldValue, stack, display)
		if err != nil {
			return err
		}
	}

	return nil
}

func showField(fieldType reflect.StructField, fieldValue reflect.Value, stack []string, display *Display) error {
	if key, ok := fieldType.Tag.Lookup("mapstructure"); ok {
		if isNotShown(key, &fieldType) {
			return nil
		}

		switch fieldValue.Kind() {

		case reflect.Ptr:
			if fieldValue.IsNil() {
				return nil
			}
			// start of a new pointer to a struct
			err := showSubStruct(fieldValue.Elem().Interface(), append(stack, key), display)
			if err != nil {
				return err
			}
			return nil

		case reflect.Struct:
			fieldIf := fieldValue.Interface()
			if _, ok := fieldIf.(fmt.Stringer); ok {
				// Pass as is. The struct has its own String() implementation
				showKeyValue(stack, display, key, fieldValue)
				return nil
			}
			var err error
			if key == ",squash" {
				// display on the same level
				err = showSubStruct(fieldIf, stack, display)
			} else {
				// start of a new struct
				err = showSubStruct(fieldIf, append(stack, key), display)
			}
			if err != nil {
				return err
			}
			return nil

		case reflect.Map:
			if fieldValue.Len() == 0 {
				return nil
			}
			if key == ",remain" {
				// special case of the map of remaining parameters: display on the same level
				showMap(stack, display, fieldValue)
				return nil
			}
			showMap(append(stack, key), display, fieldValue)
			return nil

		case reflect.Array | reflect.Slice:
			length := fieldValue.Len()
			if length > 0 && fieldValue.Index(0).Kind() == reflect.Struct {
				listStack := append(stack, key)
				for i := 0; i < length; i++ {
					if i > 0 {
						display.addKeyOnlyEntry(listStack, "-")
					}
					if err := showSubStruct(fieldValue.Index(i).Interface(), listStack, display); err != nil {
						return err
					}
				}
			} else {
				showKeyValue(stack, display, key, fieldValue)
			}
		default:
			showKeyValue(stack, display, key, fieldValue)
		}
	}
	return nil
}

func showMap(stack []string, display *Display, valueOf reflect.Value) {
	// Sort keys for a deterministic order
	keys := valueOf.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })

	for _, key := range keys {
		showKeyValue(stack, display, key.String(), valueOf.MapIndex(key))
	}
}

func showKeyValue(stack []string, display *Display, key string, valueOf reflect.Value) {
	if isNotShown(key, nil) {
		return
	}

	// This is reusing the stringify function used to build the restic flags
	convert, ok := stringify(valueOf, false)
	if ok {
		if len(convert) == 0 {
			// special case of a true flag that shows no value
			convert = append(convert, "true")
		}
		display.addEntry(stack, key, convert)
	}
}

func isNotShown(name string, fieldType *reflect.StructField) (noShow bool) {
	if name == "" || name == constants.SectionConfigurationMixinUse {
		noShow = true
	} else if fieldType != nil {
		if show, ok := fieldType.Tag.Lookup("show"); ok {
			noShow = strings.Contains(show, "noshow")
		}
	}
	return
}
