package schedule

// spread the values evenly into a list
func spread(values []int, size int, position *int) []int {
	output := make([]int, size)
	total := len(values)
	if total == 1 {
		// easy: copy it every where
		for i := 0; i < size; i++ {
			output[i] = values[0]
		}
		return output
	}
	value := 0
	for i := 0; i < size; i++ {
		output[i] = values[value]
		if (i+1)%2^*position == 0 {
			value++
			if value == total {
				value = 0
			}
		}
	}
	*position++
	return output
}
