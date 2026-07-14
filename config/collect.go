package config

import (
	"fmt"
	"reflect"
	"slices"
	"sort"

	"github.com/creativeprojects/resticprofile/util"
)

// CollectStruct walks a struct using the same tag and visibility rules as ShowStruct,
// but returns a map[string]any suitable for JSON serialization.
func CollectStruct(orig any) (map[string]any, error) {
	return collectStructValue(reflect.ValueOf(orig))
}

func collectStructValue(value reflect.Value) (map[string]any, error) {
	v, isNil := util.UnpackValue(value)
	if isNil {
		return nil, nil
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type %s, expected %s", v.Kind(), reflect.Struct)
	}
	result := make(map[string]any)
	for i := 0; i < v.Type().NumField(); i++ {
		ft, fv := v.Type().Field(i), v.Field(i)
		key, ok := fieldShown(&ft)
		if !ok {
			continue
		}
		uv, isNil := util.UnpackValue(fv)
		if isNil {
			continue
		}

		switch uv.Kind() {
		case reflect.Struct:
			if getStringer(fv) != nil {
				collectLeaf(result, key, fv)
			}
			if key == ",squash" {
				if sub, err := collectStructValue(fv); err == nil {
					for k, v := range sub {
						result[k] = v
					}
				}
			} else if sub, err := collectStructValue(fv); err == nil && len(sub) > 0 {
				result[key] = sub
			}

		case reflect.Map:
			if fv.Len() > 0 {
				m := collectMapValue(fv)
				if key == ",remain" {
					for k, v := range m {
						result[k] = v
					}
				} else {
					result[key] = m
				}
			}

		case reflect.Array, reflect.Slice:
			if isSliceWithStruct(fv) {
				if items := collectSliceValue(fv); len(items) > 0 {
					result[key] = items
				}
			} else {
				collectLeaf(result, key, fv)
			}

		default:
			collectLeaf(result, key, fv)
		}
	}
	return result, nil
}

func collectMapValue(v reflect.Value) map[string]any {
	result := make(map[string]any)
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, key := range keys {
		collectLeaf(result, key.String(), v.MapIndex(key))
	}
	return result
}

func collectSliceValue(v reflect.Value) (items []any) {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		uv, isNil := util.UnpackValue(elem)
		if isNil {
			continue
		}
		switch uv.Kind() {
		case reflect.Struct:
			if getStringer(uv) != nil {
				items = append(items, uv.Interface().(fmt.Stringer).String())
			} else if sub, err := collectStructValue(elem); err == nil && len(sub) > 0 {
				items = append(items, sub)
			}
		case reflect.Map:
			if uv.Len() > 0 {
				items = append(items, collectMapValue(uv))
			}
		case reflect.Slice, reflect.Array:
			if nested := collectSliceValue(uv); len(nested) > 0 {
				items = append(items, nested)
			}
		default:
			if vals, ok := stringify(elem, false); ok && len(vals) > 0 {
				items = append(items, vals[0])
			}
		}
	}
	return
}

func collectLeaf(result map[string]any, key string, v reflect.Value) {
	if isNotShown(key, nil) {
		return
	}
	uv, isNil := util.UnpackValue(v)
	if isNil {
		return
	}
	switch uv.Kind() {
	case reflect.Struct:
		if getStringer(uv) == nil {
			if sub, err := collectStructValue(v); err == nil && len(sub) > 0 {
				result[key] = sub
			}
		} else if vals, ok := stringify(v, false); ok && len(vals) > 0 {
			result[key] = vals[0]
		}

	case reflect.Map:
		if uv.Len() > 0 {
			result[key] = collectMapValue(uv)
		}

	case reflect.Slice, reflect.Array:
		if isSliceWithStruct(uv) {
			if items := collectSliceValue(uv); len(items) > 0 {
				result[key] = items
			}
		} else if uv.Len() > 0 {
			result[key] = collectSliceValue(uv)
		}

	default:
		if vals, ok := stringify(v, false); ok {
			switch len(vals) {
			case 0:
				result[key] = true
			case 1:
				result[key] = vals[0]
			default:
				result[key] = vals
			}
		} else if slices.Contains(allowedEmptyValueArgs, key) {
			result[key] = ""
		}
	}
}
