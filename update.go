package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/go-selfupdate/selfupdate"
	"github.com/creativeprojects/resticprofile/term"
)

func confirmAndSelfUpdate(quiet, debug bool, version string) error {
	if debug {
		selfupdate.SetLogger(clog.NewStandardLogger(clog.LevelDebug, clog.GetDefaultLogger()))
	}
	latest, found, err := selfupdate.DetectLatest("creativeprojects/resticprofile")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %v", err)
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
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %v", err)
	}
	clog.Infof("Successfully updated to version %s", latest.Version())
	return nil
}
