package config

import (
	"reflect"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// This is where it's getting hairy:
//
// Most configuration file formats allow only one declaration per section
//
// This is not the case for HCL where you can declare a bloc multiple times:
//
// "global" {
//   key1 = "value"
// }
//
// "global" {
//   key2 = "value"
// }
//
// For that matter, viper creates an slice of maps instead of a map on the other configuration file formats
//
// The code in this file deals with the slice to merge it into a single map
var (
	configOption = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		sliceOfMapsToMapHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))
)

// sliceOfMapsToMapHookFunc merges a slice of maps to a map
func sliceOfMapsToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Slice && from.Elem().Kind() == reflect.Map && (to.Kind() == reflect.Struct || to.Kind() == reflect.Map) {
			clog.Debugf("hook: from slice %+v to %+v", from.Elem(), to)
			source, ok := data.([]map[string]interface{})
			if !ok {
				return data, nil
			}
			if len(source) == 0 {
				return data, nil
			}
			if len(source) == 1 {
				return source[0], nil
			}
			// flatten the slice into one map
			convert := make(map[string]interface{})
			for _, mapItem := range source {
				for key, value := range mapItem {
					convert[key] = value
				}
			}
			return convert, nil
		}
		clog.Debugf("default from %+v to %+v", from, to)
		return data, nil
	}
}
