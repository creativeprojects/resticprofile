package maybe

import (
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// Duration implements Optional[time.Duration]
type Duration struct {
	Optional[time.Duration]
}

// SetDuration returns a maybe.Duration with value
func SetDuration(value time.Duration) Duration { return Duration{Set(value)} }

// DurationDecoder implements config parsing for maybe.Duration
func DurationDecoder() func(from, to reflect.Type, data any) (any, error) {
	fromType := reflect.TypeOf(time.Duration(0))
	valueType := reflect.TypeOf(Duration{})

	return func(from, to reflect.Type, data any) (result any, err error) {
		result = data
		if to != valueType {
			return
		}

		if value, e := cast.ToDurationE(data); e == nil {
			from = fromType
			data = value
		} else if strings.HasPrefix(e.Error(), "time:") {
			err = e
		}

		if err == nil && from == fromType {
			result = SetDuration(data.(time.Duration))
		}
		return
	}
}
