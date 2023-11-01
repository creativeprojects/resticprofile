//go:build !no_self_update

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
)

func init() {
	def := ownCommand{
		name:              "self-update",
		description:       "update to latest resticprofile",
		longDescription:   "The \"self-update\" command checks for the latest resticprofile release and updates the current application binary if a newer version is available",
		action:            selfUpdate,
		needConfiguration: false,
		flags:             map[string]string{"-q, --quiet": "update without confirmation prompt"},
	}
	ownCommands.Register([]ownCommand{
		def,
	})
	// own commands have no profile section, prevent their definition
	config.ExcludeProfileSection(def.name)
}

func selfUpdate(_ io.Writer, request commandContext) error {
	quiet := request.context.flags.quiet
	if !quiet && len(request.context.arguments) > 0 && (request.context.arguments[0] == "-q" || request.context.arguments[0] == "--quiet") {
		quiet = true
	}
	err := confirmAndSelfUpdate(quiet, request.context.flags.verbose, version, true)
	if err != nil {
		return err
	}
	return nil
}

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
