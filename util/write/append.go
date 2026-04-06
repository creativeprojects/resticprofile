package write

import "io"

// WriterAppendFunc is called for every input byte when appending it to the output buffer (is similar to buf = append(buf, byte))
type WriterAppendFunc func(dst []byte, c byte) []byte

type Append struct {
	appender WriterAppendFunc
	next     io.Writer
}

func NewAppend(w io.Writer, fn WriterAppendFunc) *Append {
	return &Append{
		appender: fn,
		next:     w,
	}
}

func (a *Append) Write(data []byte) (int, error) {
	if a.appender == nil {
		return a.next.Write(data)
	}
	dst := make([]byte, 0, len(data)*2)
	for _, c := range data {
		dst = a.appender(dst, c)
	}
	return a.next.Write(dst)
}
