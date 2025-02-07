package schtasks

import (
	"encoding/xml"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadXMLTask(t *testing.T) {
	filename := "examples/only_once.xml"
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		// no need for character conversion
		return input, nil
	}
	task := Task{}
	err = decoder.Decode(&task)
	require.NoError(t, err)

	t.Logf("%+v", task)
}

func TestSaveXMLTask(t *testing.T) {
	file, err := os.Create("output.xml")
	require.NoError(t, err)
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	task := NewTask()
	err = encoder.Encode(&task)
	require.NoError(t, err)
}
