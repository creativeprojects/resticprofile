package util

import (
	"encoding/json"
)

type Maybe[T any] struct {
	value    T
	hasValue bool
}

func Set[T any](value T) Maybe[T] {
	return Maybe[T]{
		value:    value,
		hasValue: true,
	}
}

func (m Maybe[T]) HasValue() bool {
	return m.hasValue
}

func (m Maybe[T]) Value() T {
	return m.value
}

func (m *Maybe[T]) UnmarshalJSON(data []byte) error {
	var t *T
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	if t != nil {
		*m = Set(*t)
	}

	return nil
}

func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	var t *T

	if m.hasValue {
		t = &m.value
	}

	return json.Marshal(t)
}

type MaybeBool struct {
	Maybe[bool]
}

func isTrue(maybe Maybe[bool]) bool {
	return maybe.HasValue() && maybe.Value()
}
