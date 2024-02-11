package maybe

import (
	"reflect"
	"strconv"
)

type Bool struct {
	Optional[bool]
}

func False() Bool {
	return Bool{Set(false)}
}

func True() Bool {
	return Bool{Set(true)}
}

func (value Bool) String() string {
	if !value.HasValue() {
		return ""
	}
	return strconv.FormatBool(value.Value())
}

func (value Bool) IsTrue() bool {
	return value.HasValue() && value.Value()
}

func (value Bool) IsStrictlyFalse() bool {
	return value.HasValue() && value.Value() == false
}

func (value Bool) IsFalseOrUndefined() bool {
	return !value.HasValue() || value.Value() == false
}

func (value Bool) IsUndefined() bool {
	return !value.HasValue()
}

func (value Bool) IsTrueOrUndefined() bool {
	return !value.HasValue() || value.Value() == true
}

// BoolDecoder implements config parsing for maybe.Bool
func BoolDecoder() func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	boolValueType := reflect.TypeOf(Bool{})

	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from != reflect.TypeOf(true) || to != boolValueType {
			return data, nil
		}
		boolValue, ok := data.(bool)
		if !ok {
			// it should never happen
			return data, nil
		}
		return Bool{Set(boolValue)}, nil
	}
}
