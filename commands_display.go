package main

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/ansi"
	"github.com/creativeprojects/resticprofile/util/collect"
)

func displayWriter(terminal *term.Terminal) (out func(args ...any) io.Writer, closer func()) {
	var output io.Writer = terminal
	if terminal.StdoutIsTerminal() {
		if width, _ := terminal.Size(); width > 10 {
			output = ansi.NewLineLengthWriter(terminal, width)
		}
	}

	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)

	out = func(args ...any) io.Writer {
		if len(args) > 0 {
			if msg, ok := args[0].(string); ok {
				if len(args) > 1 {
					_, _ = fmt.Fprintf(w, msg, args[1:]...)
				} else {
					_, _ = fmt.Fprint(w, msg)
				}
			}
		}
		return w
	}

	closer = func() {
		_ = w.Flush()
	}

	return
}

func getCommonUsageHelpLine(commandName string, withProfile bool) string {
	profile := ""
	if withProfile {
		profile = "[profile name.]"
	}
	return fmt.Sprintf(
		"%s [resticprofile flags] %s%s",
		ansi.Bold("resticprofile"), profile, ansi.Bold(commandName),
	)
}

func displayOwnCommands(ctx commandContext) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	for _, command := range ctx.ownCommands.commands {
		if command.hide {
			continue
		}

		out("\t%s\t%s\n", command.name, command.description)
	}
}

func displayOwnCommandHelp(ctx commandContext, commandName string) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	var command *ownCommand
	for _, c := range ctx.ownCommands.commands {
		if c.name == commandName {
			command = &c
			break
		}
	}

	if command == nil || command.hide {
		out("No help available for command \"%s\"\n", commandName)
		return
	}

	if len(command.longDescription) > 0 {
		out("%s\n\n", command.longDescription)
	} else {
		out("Purpose: %s\n\n", command.description)
	}

	commandFlags := ""
	if len(command.flags) > 0 {
		commandFlags = "[command specific flags]"
	}
	out("Usage:\n")
	out("\t%s %s\n\n", getCommonUsageHelpLine(command.name, command.needConfiguration && !command.noProfile), commandFlags)

	var flags = make([]string, 0, len(command.flags))
	for f := range command.flags {
		flags = append(flags, f)
	}
	if len(flags) > 0 {
		sort.Strings(flags)
		out("Flags:\n")
		for _, f := range flags {
			out("\t%s\t%s\n", f, command.flags[f])
		}
		out("\n")
	}
}

func displayCommonUsageHelp(ctx commandContext) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	out("resticprofile is a configuration profiles manager for backup profiles and ")
	out("is the missing link between a configuration file and restic backup\n\n")

	out("Usage:\n")
	out("\t%s [restic flags]\n", getCommonUsageHelpLine("restic-command", true))
	out("\t%s [command specific flags]\n", getCommonUsageHelpLine("resticprofile-command", true))
	out("\n")
	out(ansi.Bold("resticprofile flags:\n"))
	out(ctx.flags.usagesHelp)
	out("\n\n")
	out(ansi.Bold("resticprofile own commands:\n"))
	displayOwnCommands(ctx)
	out("\n")

	out("%s at %s\n",
		ansi.Bold("Documentation available"),
		ansi.Bold(ansi.Cyan("https://creativeprojects.github.io/resticprofile/")))
	out("\n")
}

func displayResticHelp(ctx commandContext, command string) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	resticBinary := ""
	if ctx.config != nil {
		if section, err := ctx.config.GetGlobalSection(); err == nil {
			resticBinary = section.ResticBinary
		}
	}

	if restic, err := filesearch.NewFinder().FindResticBinary(resticBinary); err == nil {
		buf := bytes.Buffer{}
		cmd := shell.NewCommand(restic, []string{"help", command})
		cmd.Stdout = &buf
		_, _, err = cmd.Run()
		if err != nil {
			out("\nFailed requesting help from restic: %s\n", err.Error())
			return
		} else if buf.Len() == 0 {
			out("\nNo help from restic for \"%s\"\n", command)
			return
		}
		replacer := strings.NewReplacer(
			// restic command => resticprofile [resticprofile flags] command
			fmt.Sprintf("restic %s", command), getCommonUsageHelpLine(command, true),
		)
		_, _ = replacer.WriteString(out(), buf.String())
	} else {
		out("restic binary not found: %s\n", err.Error())
	}

	if ctx.config != nil {
		out("\nFlags applied by resticprofile (configuration \"%s\"):\n\n", ansi.Bold(ctx.config.GetConfigFile()))

		if profileNames := ctx.config.GetProfileNames(); len(profileNames) > 0 {
			profiles := ctx.config.GetProfiles()
			sort.Strings(profileNames)
			unescaper := strings.NewReplacer(
				`\\`, `^^`,
				`\ `, ` `,
				`\'`, `'`,
				`\"`, `"`,
				`^^`, `\`,
			)

			for _, name := range profileNames {
				out("\tprofile \"%s\":", ansi.Bold(name))
				profile := profiles[name]
				cmdFlags := config.GetNonConfidentialArgs(profile, profile.GetCommandFlags(command))
				for _, flag := range cmdFlags.GetAll() {
					if strings.HasPrefix(flag, "-") {
						out("\n\t\t")
					}
					out("%s\t", ansi.Cyan(unescaper.Replace(flag)))
				}
				out("\n")
			}
		} else {
			out("none\n")
		}
	}
}

func displayHelpCommand(ctx commandContext) error {
	flags := ctx.flags

	if flags.log == "" {
		clog.GetDefaultLogger().SetHandler(clog.NewDiscardHandler()) // disable log output
	}

	validCommandName := regexp.MustCompile(`^\w{2,}[-\w]*$`).MatchString
	notHelp := collect.Not(collect.In("help"))

	helpForCommand := collect.First(flags.resticArgs, collect.With(validCommandName, notHelp))
	if helpForCommand == nil {
		helpForCommand = collect.First(flags.resticArgs, validCommandName)
	}

	if helpForCommand == nil {
		displayCommonUsageHelp(ctx)

	} else if ctx.ownCommands.Exists(*helpForCommand, true) || ctx.ownCommands.Exists(*helpForCommand, false) {
		displayOwnCommandHelp(ctx, *helpForCommand)

	} else {
		displayResticHelp(ctx, *helpForCommand)
	}

	return nil
}

func displayVersion(ctx commandContext) error {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	out("resticprofile version %s commit %s\n", ansi.Bold(version), ansi.Yellow(commit))

	// allow for the general verbose flag, or specified after the command
	arguments := ctx.request.arguments
	if ctx.flags.verbose || (len(arguments) > 0 && (arguments[0] == "-v" || arguments[0] == "--verbose")) {
		executablePath, err := util.Executable()
		if err != nil {
			executablePath = "unknown"
		}
		out("\n")
		out("\t%s:\t%s\n", "executable", executablePath)
		out("\t%s:\t%s\n", "home", "https://github.com/creativeprojects/resticprofile")
		out("\t%s:\t%s\n", "version", version)
		out("\t%s:\t%s\n", "commit", commit)
		out("\t%s:\t%s\n", "compiled", date)
		out("\t%s:\t%s\n", "by", builtBy)
		out("\t%s:\t%s\n", "go version", runtime.Version())
		out("\t%s:\t%s\n", "os", runtime.GOOS)
		out("\t%s:\t%s\n", "arch", runtime.GOARCH)
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			out("\n")
			return nil
		}
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "GOAMD64":
				out("\t%s\t%s\n", "microarchitecture", setting.Value)

			case "GOARM":
				out("\t%s\t%s\n", "arm", setting.Value)

			case "GOARM64":
				out("\t%s\t%s\n", "arm", setting.Value)
			}
		}
		out("\n")
		out("\t%s:\n", "go modules")
		for _, dep := range bi.Deps {
			out("\t\t%s\t%s\n", ansi.Cyan(dep.Path), dep.Version)
		}
		out("\n")
	}
	return nil
}

func displayProfilesCommand(ctx commandContext) error {
	displayProfiles(ctx)
	displayGroups(ctx)
	return nil
}

func displayProfiles(ctx commandContext) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	profiles := ctx.config.GetProfiles()
	keys := sortedProfileKeys(profiles)
	if len(profiles) == 0 {
		out("\nThere's no available profile in the configuration\n")
	} else {
		out("\n%s (name, sections, description):\n", ansi.Bold("Profiles available"))
		for _, name := range keys {
			sections := profiles[name].DefinedCommands()
			sort.Strings(sections)
			if len(sections) == 0 {
				out("\t%s:\t(n/a)\t%s\n", name, profiles[name].Description)
			} else {
				out("\t%s:\t(%s)\t%s\n", name, ansi.Cyan(strings.Join(sections, ", ")), profiles[name].Description)
			}
		}
	}
	out("\n")
}

func displayGroups(ctx commandContext) {
	out, closer := displayWriter(ctx.terminal)
	defer closer()

	groups := ctx.config.GetProfileGroups()
	if len(groups) == 0 {
		return
	}
	out("%s (name, profiles, description):\n", ansi.Bold("Groups available"))
	for name, groupList := range groups {
		out("\t%s:\t[%s]\t%s\n", name, ansi.Cyan(strings.Join(groupList.Profiles, ", ")), groupList.Description)
	}
	out("\n")
}

func sortedProfileKeys(data map[string]*config.Profile) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
