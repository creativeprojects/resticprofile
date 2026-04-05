package term

type nilReader struct{}

func (nilReader) Read(p []byte) (int, error) {
	return 0, nil
}
