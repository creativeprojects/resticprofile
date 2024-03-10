package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var decoders []DecoderFunc

// DecoderFunc detects a charset and provides a transformer for it.
// May return nil when no charset was detected.
// The func is responsible for setting the reader's offset (e.g. rewind) before returning.
type DecoderFunc func(reader io.ReadSeeker) transform.Transformer

// RegisterCharsetDecoder allow to register a func that detects and decodes the charset of an arbitrary text stream
func RegisterCharsetDecoder(fn DecoderFunc) {
	if fn != nil {
		decoders = append(decoders, fn)
	}
}

// NewUTF8Reader returns a reader that decodes the provided input to an UTF8 stream using registered DecoderFunc
func NewUTF8Reader(reader io.ReadSeeker) io.Reader {
	for i := len(decoders) - 1; i >= 0; i-- {
		if transformer := decoders[i](reader); transformer != nil {
			return transform.NewReader(reader, transformer)
		}
	}
	return reader
}

func utfDecoder() DecoderFunc {
	var utfBoms = [][]byte{
		{0xef, 0xbb, 0xbf}, //UTF-8
		{0xfe, 0xff},       //UTF16-BE
		{0xff, 0xfe},       //UTF16-LE
	}

	return func(reader io.ReadSeeker) transform.Transformer {
		head := make([]byte, 3)
		if headLen, err := reader.Read(head); err == nil || err == io.EOF {
			defer mustRewindToStart(reader)

			for _, bom := range utfBoms {
				bl := len(bom)
				if bl > headLen {
					continue
				}
				if bytes.Equal(bom, head[:bl]) {
					return unicode.BOMOverride(transform.Nop)
				}
			}
		}
		return nil
	}
}

func isoDecoder() DecoderFunc {
	const (
		readSize    = 4 * 1024
		maxScanSize = 128 * 1024
	)
	return func(reader io.ReadSeeker) transform.Transformer {
		if invalid, _ := isInvalidUTF(reader, readSize, maxScanSize); invalid {
			return charmap.ISO8859_1.NewDecoder()
		}
		return nil
	}
}

func isInvalidUTF(reader io.ReadSeeker, readSize, scanSize int) (invalid bool, err error) {
	defer func() {
		if err == nil {
			mustRewindToStart(reader)
		}
	}()

	buf := make([]byte, readSize)
	transformer := unicode.BOMOverride(encoding.UTF8Validator)
	validator := transform.NewReader(io.LimitReader(reader, int64(scanSize)), transformer)

	for n := 1; n > 0; {
		n, err = validator.Read(buf)
		invalid = errors.Is(err, encoding.ErrInvalidUTF8)
		if invalid || err == io.EOF {
			err = nil
			break
		}
	}
	return
}

func mustRewindToStart(seeker io.Seeker) {
	if offset, err := seeker.Seek(0, io.SeekStart); offset != 0 {
		panic(fmt.Errorf("failed reverting read offset to start: offset was %d", offset))
	} else if err != nil && err != io.EOF {
		panic(fmt.Errorf("failed reverting read offset to start: %w", err))
	}
}

func init() {
	RegisterCharsetDecoder(utfDecoder())
	RegisterCharsetDecoder(isoDecoder())
}
