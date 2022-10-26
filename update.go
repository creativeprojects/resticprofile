package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/resticprofile/term"
)

func confirmAndSelfUpdate(quiet, debug bool, version string, prerelease bool) error {
	if debug {
		selfupdate.SetLogger(clog.NewStandardLogger(clog.LevelDebug, clog.GetDefaultLogger()))
	}
	updater, _ := selfupdate.NewUpdater(
		selfupdate.Config{
			Validator:  &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
			Prerelease: prerelease,
		})
	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.NewRepositorySlug("creativeprojects", "resticprofile"))
	if err != nil {
		return fmt.Errorf("unable to detect latest version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		clog.Infof("Current version (%s) is the latest", version)
		return nil
	}

	// don't ask in quiet mode
	if !quiet && !term.AskYesNo(os.Stdin, fmt.Sprintf("Do you want to update to version %s", latest.Version()), true) {
		fmt.Println("Never mind")
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return fmt.Errorf("unable to update binary: %w", err)
	}
	clog.Infof("Successfully updated to version %s", latest.Version())
	return nil
}
