package restic

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	sup "github.com/creativeprojects/go-selfupdate"
)

const (
	VersionLatest = "latest"
	Executable    = "restic"
	owner         = "restic"
	repo          = "restic"
	minVersion    = "0.9.4"
	checksumAsset = "SHA256SUMS"
)

var (
	githubSource   sup.Source
	defaultUpdater *sup.Updater
)

func init() {
	githubSource = newUpdateSource()
	defaultUpdater = newUpdater("", "", readResticPGPKey(), githubSource)
}

func newUpdateSource() sup.Source {
	if source, err := sup.NewGitHubSource(sup.GitHubConfig{}); err == nil {
		// Using a cached source as we may retry finding the correct version (this saves some API calls)
		return newCachedSource(owner, repo, source)
	} else {
		panic(err)
	}
}

func newUpdater(os, arch string, key []byte, source sup.Source) *sup.Updater {
	updater, err := sup.NewUpdater(sup.Config{
		Source:    source,
		Validator: sup.NewChecksumWithPGPValidator(checksumAsset, key),
		OS:        os,
		Arch:      arch,
	})
	if err != nil {
		panic(err)
	}
	return updater
}

var versionCommandPattern = regexp.MustCompile("restic ([\\d.]+)[ -].+")

// GetVersion returns the version of the executable
func GetVersion(executable string) (string, error) {
	cmd := exec.Command(executable, "version")
	if output, err := cmd.Output(); err == nil {
		if match := versionCommandPattern.FindSubmatch(output); match != nil {
			return string(match[1]), nil
		}
		return "", fmt.Errorf("restic returned no valid version: %s", strings.TrimSpace(string(output)))
	} else {
		return "", err
	}
}

// DownloadBinary downloads a specific restic binary to executable for the current platform.
// Version can be empty or "latest" to download the latest available restic binary for the current platform.
func DownloadBinary(executable, version string) error {
	ctx, closer := context.WithTimeout(context.Background(), time.Minute*15)
	defer closer()
	return download(ctx, defaultUpdater, executable, version, minVersion)
}

func download(ctx context.Context, updater *sup.Updater, executable, version, minVersion string) error {
	if version == VersionLatest {
		version = ""
	}

	// Check input and create empty executable if missing
	if s, err := os.Stat(executable); os.IsNotExist(err) {
		if file, err := os.OpenFile(executable, os.O_CREATE, 0755); err == nil {
			_ = file.Close()
			defer func() {
				if s, e := os.Stat(executable); e == nil && s.Size() == 0 && !s.IsDir() {
					_ = os.Remove(executable)
				}
			}()
		} else {
			return err
		}
	} else if err == nil && s.IsDir() {
		return fmt.Errorf("%s is a directory", executable)
	}

	slug := sup.NewRepositorySlug(owner, repo)

	release, found, err := updater.DetectVersion(ctx, slug, version)
	if !found && err == nil && !strings.HasPrefix(version, "v") {
		release, found, err = updater.DetectVersion(ctx, slug, fmt.Sprintf("v%s", version))
	}
	if !found && err == nil {
		err = fmt.Errorf(`restic version "%s" not found`, version)
	}
	if found && err == nil && len(minVersion) > 0 && release.LessThan(minVersion) {
		err = fmt.Errorf(`restic version "%s" is less than %s`, release.Version(), minVersion)
	}
	if err != nil {
		return err
	}

	return updater.UpdateTo(ctx, release, executable)
}
