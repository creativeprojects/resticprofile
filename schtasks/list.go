//go:build windows

package schtasks

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

const (
	commonListSeparator = ":  "
	otherListSeparator  = ": "
	listFolderKey       = "Folder"
)

func getTaskInfoFromList(input io.Reader) ([]map[string]string, error) {
	currentFolder := ""
	output := make([]map[string]string, 0, 10)
	record := make(map[string]string, 30)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			if len(record) > 0 {
				record[listFolderKey] = currentFolder
				output = append(output, record)
				record = make(map[string]string, 30)
			}
			continue
		}
		key, value, found := strings.Cut(line, commonListSeparator)
		if !found {
			// if the line doesn't contain a colon followed by 2 spaces, it means it's either "Folder:" or the longest key.
			key, value, found = strings.Cut(line, otherListSeparator)
			if !found {
				return output, errors.New("invalid line format: " + line)
			}
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(key) == 0 || len(value) == 0 {
			continue
		}
		if key == listFolderKey {
			currentFolder = value
			continue
		}
		if _, exists := record[key]; exists {
			return output, errors.New("duplicate key found in task info: " + key)
		}
		record[key] = value
	}
	if err := scanner.Err(); err != nil {
		return output, err
	}
	if len(record) > 0 {
		record[listFolderKey] = currentFolder
		output = append(output, record)
	}
	return output, nil
}

func getFirstField(data []map[string]string, fieldName string) string {
	for _, record := range data {
		if value, exists := record[fieldName]; exists {
			return value
		}
	}
	return ""
}
