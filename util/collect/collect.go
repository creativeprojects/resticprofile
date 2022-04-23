package collect

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
