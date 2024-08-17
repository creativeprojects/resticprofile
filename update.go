//go:build !no_self_update

package main

import (
	"io"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/update"
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

func selfUpdate(_ io.Writer, ctx commandContext) error {
	quiet := ctx.flags.quiet
	if !quiet && len(ctx.request.arguments) > 0 && (ctx.request.arguments[0] == "-q" || ctx.request.arguments[0] == "--quiet") {
		quiet = true
	}
	err := update.ConfirmAndSelfUpdate(quiet, ctx.flags.verbose, version, true)
	if err != nil {
		return err
	}
	return nil
}
