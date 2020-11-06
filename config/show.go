package config

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

const (
	templateIndent = "    " // 4 spaces
)

// ShowStruct write out to w a human readable text representation of the orig parameter
func ShowStruct(w io.Writer, orig interface{}) error {
	return showSubStruct(w, orig, "")
}

func showSubStruct(outputWriter io.Writer, orig interface{}, prefix string) error {
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

	// make a temporary buffer to display sub struct and map after direct properties (which are buffered through a tabWriter)
	buffer := bufio.NewWriter(outputWriter)
	tabWriter := tabwriter.NewWriter(outputWriter, 0, 2, 2, ' ', 0)
	prefix = addIndentation(prefix)

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
				fmt.Fprintf(buffer, "%s%s:\n", prefix, key)
				err := showSubStruct(buffer, valueOf.Field(i).Elem().Interface(), prefix)
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Struct {
				// start of a new struct
				fmt.Fprintf(buffer, "%s%s:\n", prefix, key)
				err := showSubStruct(buffer, valueOf.Field(i).Interface(), prefix)
				if err != nil {
					return err
				}
				continue
			}
			if valueOf.Field(i).Kind() == reflect.Map {
				if valueOf.Field(i).Len() == 0 {
					continue
				}
				// special case of the map of remaining parameters...
				if key == ",remain" {
					showMap(tabWriter, prefix, valueOf.Field(i))
					continue
				}
				fmt.Fprintf(buffer, "%s%s:\n", prefix, key)
				showNewMap(buffer, prefix, valueOf.Field(i))
				continue
			}
			showKeyValue(tabWriter, prefix, key, valueOf.Field(i))
		}
	}

	tabWriter.Flush()
	fmt.Fprintln(buffer, "")
	buffer.Flush()

	return nil
}

func addIndentation(indent string) string {
	return indent + templateIndent
}

func showNewMap(outputWriter io.Writer, prefix string, valueOf reflect.Value) {
	subWriter := tabwriter.NewWriter(outputWriter, 0, 2, 2, ' ', 0)
	prefix = addIndentation(prefix)
	showMap(subWriter, prefix, valueOf)
	subWriter.Flush()
	fmt.Fprintln(outputWriter, "")
}

func showMap(tabWriter io.Writer, prefix string, valueOf reflect.Value) {
	iter := valueOf.MapRange()
	for iter.Next() {
		showKeyValue(tabWriter, prefix, iter.Key().String(), iter.Value())
	}
}

func showKeyValue(tabWriter io.Writer, prefix, key string, valueOf reflect.Value) {
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
		if len(convert) > 0 {
			fmt.Fprintf(tabWriter, "%s%s:\t%s\n", prefix, key, convert[0])
		}
		if len(convert) > 1 {
			for i := 1; i < len(convert); i++ {
				fmt.Fprintf(tabWriter, "%s\t%s\n", prefix, convert[i])
			}
		}
	}
}
