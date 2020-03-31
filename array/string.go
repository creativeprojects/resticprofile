package array

// FindString searches for a needle in a haystack. If not found, it returns (-1, false)
func FindString(haystack []string, needle string) (int, bool) {
	if haystack == nil || len(haystack) == 0 {
		return -1, false
	}
	for index, value := range haystack {
		if value == needle {
			return index, true
		}
	}
	return -1, false
}
