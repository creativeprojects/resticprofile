package ansi

// RunesLength returns the visible content length (and index at the length)
func RunesLength(src []rune, maxLength int) (length, index int) {
	esc := rune(Escape[0])
	inEsc := false
	index = -1
	for i, r := range src {
		if r == esc {
			inEsc = true
		} else if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
		} else {
			if length == maxLength {
				index = i
			}
			length++
		}
	}
	if index < 0 {
		index = len(src)
	}
	return
}
