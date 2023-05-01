package term

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type askYesNoTestData struct {
	input         string
	defaultAnswer bool
	expected      bool
}

func TestAskYesNo(t *testing.T) {
	testData := []askYesNoTestData{
		// Empty answer => will follow the defaultAnswer
		{"", true, true},
		{"", false, false},
		{"\n", true, true},
		{"\n", false, false},
		{"\r\n", true, true},
		{"\r\n", false, false},
		// Garbage answer => will always return false
		{"aa", true, false},
		{"aa", false, false},
		{"aa\n", true, false},
		{"aa\n", false, false},
		{"aa\r\n", true, false},
		{"aa\r\n", false, false},
		// Answer yes
		{"y", true, true},
		{"y", false, true},
		{"y\n", true, true},
		{"y\n", false, true},
		{"y\r\n", true, true},
		{"y\r\n", false, true},
		// Full answer yes
		{"yes", true, true},
		{"yes", false, true},
		{"yes\n", true, true},
		{"yes\n", false, true},
		{"yes\r\n", true, true},
		{"yes\r\n", false, true},
		// Answer no
		{"n", true, false},
		{"n", false, false},
		{"n\n", true, false},
		{"n\n", false, false},
		{"n\r\n", true, false},
		{"n\r\n", false, false},
		// Full answer no
		{"no", true, false},
		{"no", false, false},
		{"no\n", true, false},
		{"no\n", false, false},
		{"no\r\n", true, false},
		{"no\r\n", false, false},
	}
	for _, testItem := range testData {
		StartRecording(RecordOutput)
		result := AskYesNo(
			bytes.NewBufferString(testItem.input),
			"message",
			testItem.defaultAnswer,
		)
		assert.Contains(t, StopRecording(), "message? (")
		assert.Equal(t, testItem.expected, result, "when input was %q", testItem.input)
	}
}

func TestExamplePrint(t *testing.T) {
	SetOutput(os.Stdout)
	Print("ExamplePrint")
	// Output: ExamplePrint
}

func TestOutputCapture(t *testing.T) {
	SetOutput(os.Stdout)
	SetErrorOutput(os.Stderr)

	StartRecording(RecordOutput)
	_, _ = Print("ExamplePrint")
	assert.Equal(t, "ExamplePrint", ReadRecording())
	_, _ = Print("1")
	assert.Equal(t, "1", StopRecording())

	StartRecording(RecordError)
	_, _ = fmt.Fprintf(GetErrorOutput(), "ExamplePrint")
	assert.Equal(t, "ExamplePrint", StopRecording())

	StartRecording(RecordBoth)
	_, _ = fmt.Fprintf(GetOutput(), "Info")
	_, _ = fmt.Fprintf(GetErrorOutput(), "Error")
	assert.Equal(t, "InfoError", StopRecording())
}

func TestCanRedirectTermOutput(t *testing.T) {
	defer SetOutput(os.Stdout)
	message := "TestCanRedirectTermOutput"
	buffer := &bytes.Buffer{}
	SetOutput(buffer)
	_, err := Print(message)
	assert.NoError(t, err)
	assert.Equal(t, message, buffer.String())
}
