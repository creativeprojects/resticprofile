package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
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
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	v := semver.MustParse(version)
	if latest.Version.LTE(v) {
		clog.Infof("Current version (%s) is the latest", version)
		return nil
	}

	if !askYesNo(os.Stdin, fmt.Sprint("Do you want to update to version ", latest.Version), true) {
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
	clog.Infof("Successfully updated to version %s", latest.Version)
	return nil
}

func askYesNo(reader io.Reader, message string, defaultAnswer bool) bool {
	if !strings.HasSuffix(message, "?") {
		message += "?"
	}
	question := ""
	input := ""
	if defaultAnswer {
		question = "(Y/n)"
		input = "y"
	} else {
		question = "(y/N)"
		input = "n"
	}
	fmt.Printf("%s %s: ", message, question)
	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		input = strings.TrimSpace(strings.ToLower(scanner.Text()))
		if len(input) > 1 {
			// take only the first character
			input = input[:1]
		}
	}

	if input == "" {
		return defaultAnswer
	}
	if input == "y" {
		return true
	}
	return false
}
