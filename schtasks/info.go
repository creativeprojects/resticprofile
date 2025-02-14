package schtasks

import (
	"bytes"
	"encoding/csv"
	"io"
)

func getTaskInfo(taskName string) ([][]string, error) {
	buffer := &bytes.Buffer{}
	err := readTaskInfo(taskName, buffer)
	if err != nil {
		return nil, err
	}
	output, err := getCSV(buffer)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func getCSV(input io.Reader) ([][]string, error) {
	reader := csv.NewReader(input)
	return reader.ReadAll()
}
