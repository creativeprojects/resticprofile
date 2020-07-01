package schedule

import (
	"path"
	"path/filepath"
)

// absolutePathToBinary returns an absolute path to the resticprofile binary
func absolutePathToBinary(currentDir, binaryPath string) string {
	binary := binaryPath
	if !filepath.IsAbs(binary) {
		binary = path.Join(currentDir, binary)
	}
	return binary
}
