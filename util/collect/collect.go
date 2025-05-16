package collect

import "slices"

// All collects all items from input that satisfy the condition
// Empty or nil input returns nil output
func All[I ~[]T, T any](input I, condition func(t T) bool) (output []T) {
	for _, item := range input {
		if condition(item) {
			output = append(output, item)
		}
	}
	return
}

// Not inverts the meaning of condition
func Not[T any, C func(t T) bool](condition C) C {
	not := func(t T) bool { return !condition(t) }
	return not
}

// In returns a new condition that is true as one of the values matches
func In[E comparable](values ...E) (condition func(item E) bool) {
	return func(item E) bool { return slices.Contains(values, item) }
}

// With returns a new condition that is true when all conditions match
func With[C ~func(t T) bool, T any](conditions ...C) (condition C) {
	return func(item T) bool {
		for _, c := range conditions {
			if !c(item) {
				return false
			}
		}
		return true
	}
}

// From translates a slice into another using a mapper func (T) => (R).
// Empty or nil input returns nil output
func From[I ~[]T, T, R any](input I, mapper func(t T) R) (output []R) {
	if len(input) > 0 {
		output = make([]R, len(input))
		for index, item := range input {
			output[index] = mapper(item)
		}
	}
	return
}

// Filter a slice using a filter func (T) => (bool).
func Filter[I ~[]T, T any](input I, filterFunc func(t T) bool) (output []T) {
	if len(input) > 0 {
		output = make([]T, 0, len(input))
		for _, item := range input {
			if filterFunc(item) {
				output = append(output, item)
			}
		}
	}
	return
}
