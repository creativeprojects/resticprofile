package config

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util"
)

// ShowStruct write out to w a human-readable text representation of the orig parameter
func ShowStruct(w io.Writer, orig any, name string) error {
	display := newDisplay(name, w)
	err := showSubStruct([]string{}, display, orig)
	if err != nil {
		return err
	}
	display.Flush()
	return nil
}

func showSubStruct(stack []string, display *Display, orig any) (err error) {
	return showSubStructValue(stack, display, reflect.ValueOf(orig))
}

func showSubStructValue(stack []string, display *Display, value reflect.Value) (err error) {
	valueOf, isNil := util.UnpackValue(value)
	if isNil {
		err = nil
	} else if valueOf.Kind() != reflect.Struct {
		err = fmt.Errorf("unsupported type %s, expected %s", valueOf.Kind(), reflect.Struct)
	} else {
		typeOf := valueOf.Type()
		for i := 0; i < typeOf.NumField() && err == nil; i++ {
			fieldType := typeOf.Field(i)
			fieldValue := valueOf.Field(i)
			err = showField(stack, display, &fieldType, fieldValue)
		}
	}
	return
}

func fieldShown(fieldType *reflect.StructField) (key string, show bool) {
	for _, tagName := range []string{"mapstructure", "show"} {
		if key, show = fieldType.Tag.Lookup(tagName); show {
			show = !isNotShown(key, fieldType)
			break
		}
	}
	return
}

func showField(stack []string, display *Display, fieldType *reflect.StructField, fieldValue reflect.Value) (err error) {
	if key, ok := fieldShown(fieldType); ok {

		value, isNil := util.UnpackValue(fieldValue)
		if isNil {
			return
		}

		switch value.Kind() {
		case reflect.Struct:
			if getStringer(fieldValue) != nil {
				// Pass as is. The struct has its own String() implementation
				err = showKeyValue(stack, display, key, fieldValue)
				if err != nil {
					break
				}
			}

			if key == ",squash" {
				// display on the same level
				err = showSubStructValue(stack, display, fieldValue)
			} else {
				// start of a new struct
				err = showSubStructValue(append(stack, key), display, fieldValue)
			}

		case reflect.Map:
			if fieldValue.Len() > 0 {
				if key == ",remain" {
					// special case of the map of remaining parameters: display on the same level
					err = showMap(stack, display, fieldValue)
				} else {
					err = showMap(append(stack, key), display, fieldValue)
				}
			}

		case reflect.Array | reflect.Slice:
			if isSliceWithStruct(fieldValue) {
				err = showSliceWithStruct(stack, display, key, fieldValue)
			} else {
				err = showKeyValue(stack, display, key, fieldValue)
			}

		default:
			err = showKeyValue(stack, display, key, fieldValue)
		}
	}
	return
}

// isSliceWithStruct returns true if the first element is a struct that does not implement fmt.Stringer
func isSliceWithStruct(fieldValue reflect.Value) bool {
	if fieldValue.Kind() == reflect.Array || fieldValue.Kind() == reflect.Slice {
		if length := fieldValue.Len(); length > 0 {
			value, _ := util.UnpackValue(fieldValue.Index(0))
			return value.Kind() == reflect.Struct && getStringer(value) == nil
		}
	}
	return false
}

func showSliceWithStruct(stack []string, display *Display, key string, fieldValue reflect.Value) (err error) {
	listStack := append(stack, key)
	for i, length := 0, fieldValue.Len(); i < length; i++ {
		if i > 0 {
			display.addKeyOnlyEntry(listStack, "-")
		}
		value := fieldValue.Index(i)
		if err = showSubStructValue(listStack, display, value); err != nil {
			break
		}
	}
	return
}

func showMap(stack []string, display *Display, valueOf reflect.Value) (err error) {
	// Sort keys for a deterministic order
	keys := valueOf.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })

	for _, key := range keys {
		name := key.String()
		if err = showKeyValue(stack, display, name, valueOf.MapIndex(key)); err != nil {
			break
		}
	}
	return
}

func showKeyValue(stack []string, display *Display, key string, valueOf reflect.Value) (err error) {
	if isNotShown(key, nil) {
		return nil
	}

	value, isNil := util.UnpackValue(valueOf)
	if isNil {
		return nil
	}

	if value.Kind() == reflect.Struct && getStringer(value) == nil {
		err = showSubStructValue(append(stack, key), display, valueOf)

	} else if isSliceWithStruct(value) {
		err = showSliceWithStruct(stack, display, key, value)

	} else {
		// This is reusing the stringify function used to build the restic flags
		if convert, ok := stringify(valueOf, false); ok {
			if len(convert) == 0 {
				// special case of a true flag that shows no value
				convert = append(convert, "true")
			}
			display.addEntry(stack, key, convert)
		} else if slices.Contains(allowedEmptyValueArgs, key) {
			// special case of an empty string that needs to be shown
			display.addEntry(stack, key, []string{`""`})
		}
	}
	return err
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
