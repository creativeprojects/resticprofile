package write

import "io"

type mockWriteCloser struct {
	writer func(data []byte) (int, error)
	closer func() error
}

func newMockWriteCloser(writer func(data []byte) (int, error), closer func() error) mockWriteCloser {
	return mockWriteCloser{
		writer: writer,
		closer: closer,
	}
}

func (m mockWriteCloser) Write(data []byte) (int, error) {
	return m.writer(data)
}

func (m mockWriteCloser) Close() error {
	return m.closer()
}

var _ io.WriteCloser = &Append{}
