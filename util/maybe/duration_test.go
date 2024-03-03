package maybe_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
)

func TestDurationDecoder(t *testing.T) {
	fixtures := []struct {
		toUnexpected bool
		source       any
		expected     any
	}{
		// same value returned as the "to" type in unexpected
		{
			toUnexpected: true,
			source:       "anything",
			expected:     "anything",
		},
		// already at target type
		{
			source:   maybe.SetDuration(5 * time.Second),
			expected: maybe.SetDuration(5 * time.Second),
		},
		// convert from duration
		{
			source:   15 * time.Second,
			expected: maybe.SetDuration(15 * time.Second),
		},
		// convert from number is unexpected
		{
			source:   int(25 * time.Second),
			expected: maybe.SetDuration(25 * time.Second),
		},
		// convert from string
		{
			source:   "32m60s",
			expected: maybe.SetDuration(33 * time.Minute),
		},
		// convert from empty string
		{
			source:   "",
			expected: errors.New(`time: invalid duration "ns"`),
		},
		// string parse error
		{
			source:   "invalid",
			expected: errors.New(`time: invalid duration "invalid"`),
		},
	}
	for index, fixture := range fixtures {
		decoder := maybe.DurationDecoder()

		t.Run(fmt.Sprintf("%d", index), func(t *testing.T) {
			to := reflect.TypeOf(maybe.Duration{})
			if fixture.toUnexpected {
				to = reflect.TypeOf(false)
			}
			from := reflect.TypeOf(fixture.source)

			decoded, err := decoder(from, to, fixture.source)
			if fe, ok := fixture.expected.(error); ok {
				assert.Equal(t, fe, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, fixture.expected, decoded)
			}
		})
	}
}

func TestDurationJSON(t *testing.T) {
	fixtures := []struct {
		source   maybe.Duration
		expected string
	}{
		{
			source:   maybe.Duration{},
			expected: "null",
		},
		{
			source:   maybe.SetDuration(0),
			expected: "0",
		},
		{
			source:   maybe.SetDuration(25 * time.Minute),
			expected: strconv.Itoa(int(25 * time.Minute)),
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.source.String(), func(t *testing.T) {
			// encode value into JSON
			encoded, err := fixture.source.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, string(encoded))

			// decode value from JSON
			decodedValue := maybe.Duration{}
			err = decodedValue.UnmarshalJSON(encoded)
			assert.NoError(t, err)
			assert.Equal(t, fixture.source, decodedValue)
		})
	}
}

func TestDurationString(t *testing.T) {
	fixtures := []struct {
		source   maybe.Duration
		expected string
	}{
		{source: maybe.Duration{}, expected: ""},
		{source: maybe.SetDuration(0), expected: "0s"},
		{source: maybe.SetDuration(5 * time.Minute), expected: "5m0s"},
		{source: maybe.SetDuration(-10 * time.Minute), expected: "-10m0s"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.source.String(), func(t *testing.T) {
			assert.Equal(t, fixture.expected, fixture.source.String())
		})
	}
}
