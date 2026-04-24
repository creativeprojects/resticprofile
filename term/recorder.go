package term

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type Recorder struct {
	inputWriter *os.File
	inputReader *os.File
	wg          sync.WaitGroup
}

func NewRecorder(output io.Writer) (*Recorder, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("cannot create OS pipe: %w", err)
	}
	recorder := &Recorder{
		inputReader: r,
		inputWriter: w,
	}
	recorder.wg.Go(func() {
		_, _ = io.Copy(output, recorder.inputReader)
	})
	return recorder, nil
}

func (recorder *Recorder) Close() error {
	var err error
	err = errors.Join(err, recorder.inputWriter.Close())
	recorder.wg.Wait()
	// now that we finished reading, we can close the reader side
	err = errors.Join(err, recorder.inputReader.Close())
	return err
}
