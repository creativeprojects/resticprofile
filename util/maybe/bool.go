package maybe

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/spf13/cast"
)

type Bool struct {
	Optional[bool]
}

func SetBool(value bool) Bool { return Bool{Set(value)} }

func UnsetBool() Bool { return Bool{} }

func False() Bool { return SetBool(false) }

func True() Bool { return SetBool(true) }

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

func BoolFromNilable(value *bool) Bool {
	if value == nil {
		return UnsetBool()
	}
	return SetBool(*value)
}

// BoolDecoder implements config parsing for maybe.Bool
func BoolDecoder() func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	fromType := reflect.TypeOf(true)
	valueType := reflect.TypeOf(Bool{})

	return func(from reflect.Type, to reflect.Type, data interface{}) (result interface{}, err error) {
		result = data
		if to != valueType {
			return
		}

		if value, e := cast.ToBoolE(data); e == nil {
			from = fromType
			data = value
		} else if errors.Is(e, new(strconv.NumError)) {
			err = e
		}

		if err == nil && from == fromType {
			result = SetBool(data.(bool))
		}
		return
	}
}
