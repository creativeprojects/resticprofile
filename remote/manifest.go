package remote

import (
	"bytes"
	"strings"
)

const (
	ManifestFilename             = "__MANIFEST__"
	ManifestKeyVersion           = "Version"
	ManifestKeyConfigurationFile = "ConfigurationFile"
	ManifestKeyProfileName       = "ProfileName"
	ManifestKeyCommandLine       = "CommandLine"
)

var (
	manifestParameterSeparator = []byte(":")
	manifestLineSeparator      = []byte("\n")
)

// Create a manifest file with the given parameters
func CreateManifest(params map[string]string) []byte {
	buf := &bytes.Buffer{}
	for key, value := range params {
		buf.WriteString(key)
		buf.Write(manifestParameterSeparator)
		buf.WriteString(value)
		buf.Write(manifestLineSeparator)
	}
	return buf.Bytes()
}

func ParseManifest(data []byte) map[string]string {
	lines := bytes.Split(data, manifestLineSeparator)
	result := make(map[string]string, len(lines))
	for _, line := range lines {
		key, value, found := bytes.Cut(line, manifestParameterSeparator)
		if found {
			result[string(key)] = strings.TrimSpace(string(value))
		}
	}
	return result
}
