package collect

// This file contains additional implementations that are currently only used in tests
// They may be removed in the future if it turns out that the implementations are not useful

// First collects the first item from input that satisfies the condition
// Empty or nil input returns nil output
func First[I ~[]T, T any](input I, condition func(t T) bool) *T {
	for _, item := range input {
		if condition(item) {
			return &item
		}
	}
	return nil
}

// Last collects the last item from input that satisfies the condition
// Empty or nil input returns nil output
func Last[I ~[]T, T any](input I, condition func(t T) bool) *T {
	for i := len(input) - 1; i >= 0; i-- {
		if condition(input[i]) {
			return &input[i]
		}
	}
	return nil
}

// FromMap translates a map into another using a mapper func (k1, v2) => (k2, v2, include).
// Empty or nil input returns nil output
func FromMap[M ~map[K1]V1, K1, K2 comparable, V1, V2 any](input M, mapper func(K1, V1) (K2, V2, bool)) (output map[K2]V2) {
	if len(input) > 0 {
		output = make(map[K2]V2, len(input))
		for k1, v1 := range input {
			k2, v2, include := mapper(k1, v1)
			if include {
				output[k2] = v2
			}
		}
	}
	return
}
