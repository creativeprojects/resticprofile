package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func confirmAndSelfUpdate(debug bool) error {
	if debug {
		selfupdate.EnableLog()
	}
	latest, found, err := selfupdate.DetectLatest("creativeprojects/resticprofile")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %v", err)
	}
	if !found {
		return fmt.Errorf("latest version could not be found from github repository")
	}

	v := semver.MustParse(version)
	if latest.Version.LTE(v) {
		clog.Infof("Current version (%s) is the latest", version)
		return nil
	}

	fmt.Print("Do you want to update to version ", latest.Version, "? (Y/n): ")
	input := "n"
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = strings.ToLower(scanner.Text())
	}
	if input == "n" {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %v", err)
	}
	clog.Infof("Successfully updated to version %s", latest.Version)
	return nil
}
