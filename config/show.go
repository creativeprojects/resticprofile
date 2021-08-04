package config

import (
	"fmt"
	"io"
	"reflect"
)

// ShowStruct write out to w a human readable text representation of the orig parameter
func ShowStruct(w io.Writer, orig interface{}, name string) error {
	display := newDisplay(name, w)
	err := showSubStruct(orig, []string{}, display)
	if err != nil {
		return nil
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
		field := typeOf.Field(i)

		if key, ok := field.Tag.Lookup("mapstructure"); ok {
			if key == "" {
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Ptr {
				if valueOf.Field(i).IsNil() {
					continue
				}
				// start of a new pointer to a struct
				err := showSubStruct(valueOf.Field(i).Elem().Interface(), append(stack, key), display)
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Struct {
				fieldIf := valueOf.Field(i).Interface()
				if _, ok := fieldIf.(fmt.Stringer); ok {
					// Pass as is. The struct has its own String() implementation
				} else {
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
					continue
				}
			}
			if valueOf.Field(i).Kind() == reflect.Map {
				if valueOf.Field(i).Len() == 0 {
					continue
				}
				if key == ",remain" {
					// special case of the map of remaining parameters: display on the same level
					showMap(stack, display, valueOf.Field(i))
					continue
				}
				showMap(append(stack, key), display, valueOf.Field(i))
				continue
			}
			showKeyValue(stack, display, key, valueOf.Field(i))
		}
	}

	return nil
}

func showMap(stack []string, display *Display, valueOf reflect.Value) {
	iter := valueOf.MapRange()
	for iter.Next() {
		showKeyValue(stack, display, iter.Key().String(), iter.Value())
	}
}

func showKeyValue(stack []string, display *Display, key string, valueOf reflect.Value) {
	// hard-coded case for "inherit": we don't need to display it
	if key == "inherit" {
		return
	}
	// This is reusing the stringifyValue function used to build the restic flags
	convert, ok := stringifyValue(valueOf)
	if ok {
		if len(convert) == 0 {
			// special case of a true flag that shows no value
			convert = append(convert, "true")
		}
		display.addEntry(stack, key, convert)
	}
}
