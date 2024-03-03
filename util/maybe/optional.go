package maybe

import (
	"encoding/json"
	"fmt"
)

type Optional[T any] struct {
	value    T
	hasValue bool
}

func Set[T any](value T) Optional[T] {
	return Optional[T]{
		value:    value,
		hasValue: true,
	}
}

func (m Optional[T]) HasValue() bool {
	return m.hasValue
}

func (m Optional[T]) Value() T {
	return m.value
}

func (m Optional[T]) String() string {
	if !m.HasValue() {
		return ""
	}
	return fmt.Sprintf("%v", m.Value())
}

func (m *Optional[T]) UnmarshalJSON(data []byte) error {
	var t *T
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	if t != nil {
		*m = Set(*t)
	}

	return nil
}

func (m Optional[T]) MarshalJSON() ([]byte, error) {
	var t *T

	if m.hasValue {
		t = &m.value
	}

	return json.Marshal(t)
}
