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
	"github.com/creativeprojects/resticprofile/util/ansi"
	"github.com/creativeprojects/resticprofile/util/collect"
)

func displayWriter(output io.Writer) (out func(args ...any) io.Writer, closer func()) {
	if term.GetOutput() == output {
		output = term.GetColorableOutput()

		if width, _ := term.OsStdoutTerminalSize(); width > 10 {
			output = ansi.NewLineLengthWriter(output, width)
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

func displayOwnCommands(output io.Writer, request commandRequest) {
	out, closer := displayWriter(output)
	defer closer()

	for _, command := range request.ownCommands.commands {
		if command.hide {
			continue
		}

		out("\t%s\t%s\n", command.name, command.description)
	}
}

func displayOwnCommandHelp(output io.Writer, commandName string, request commandRequest) {
	out, closer := displayWriter(output)
	defer closer()

	var command *ownCommand
	for _, c := range request.ownCommands.commands {
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
	out("\t%s %s\n\n", getCommonUsageHelpLine(command.name, command.needConfiguration), commandFlags)

	var flags []string
	for f, _ := range command.flags {
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

func displayCommonUsageHelp(output io.Writer, request commandRequest) {
	out, closer := displayWriter(output)
	defer closer()

	out("resticprofile is a configuration profiles manager for backup profiles and ")
	out("is the missing link between a configuration file and restic backup\n\n")

	out("Usage:\n")
	out("\t%s [restic flags]\n", getCommonUsageHelpLine("restic-command", true))
	out("\t%s [command specific flags]\n", getCommonUsageHelpLine("resticprofile-command", true))
	out("\n")
	out(ansi.Bold("resticprofile flags:\n"))
	out(request.flags.usagesHelp)
	out("\n\n")
	out(ansi.Bold("resticprofile own commands:\n"))
	displayOwnCommands(out(), request)
	out("\n")

	out("%s at %s\n",
		ansi.Bold("Documentation available"),
		ansi.Bold(ansi.Cyan("https://creativeprojects.github.io/resticprofile/")))
	out("\n")
}

func displayResticHelp(output io.Writer, configuration *config.Config, flags commandLineFlags, command string) {
	out, closer := displayWriter(output)
	defer closer()

	// try to load the config
	if configuration == nil {
		if file, err := filesearch.FindConfigurationFile(flags.config); err == nil {
			if configuration, err = config.LoadFile(file, flags.format); err != nil {
				configuration = nil
			}
		}
	}

	resticBinary := ""
	if configuration != nil {
		if section, err := configuration.GetGlobalSection(); err == nil {
			resticBinary = section.ResticBinary
		}
	}

	if restic, err := filesearch.FindResticBinary(resticBinary); err == nil {
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

	if configuration != nil {
		out("\nFlags applied by resticprofile (configuration \"%s\"):\n\n", ansi.Bold(configuration.GetConfigFile()))

		if profileNames := configuration.GetProfileNames(); len(profileNames) > 0 {
			profiles := configuration.GetProfiles()
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

func displayHelpCommand(output io.Writer, request commandRequest) error {
	flags := request.flags

	out, closer := displayWriter(output)
	defer closer()

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
		displayCommonUsageHelp(out("\n"), request)

	} else if request.ownCommands.Exists(*helpForCommand, true) || request.ownCommands.Exists(*helpForCommand, false) {
		displayOwnCommandHelp(out("\n"), *helpForCommand, request)

	} else {
		displayResticHelp(out(), request.config, flags, *helpForCommand)
	}

	return nil
}

func displayVersion(output io.Writer, request commandRequest) error {
	out, closer := displayWriter(output)
	defer closer()

	out("resticprofile version %s commit %s\n", ansi.Bold(version), ansi.Yellow(commit))

	// allow for the general verbose flag, or specified after the command
	if request.flags.verbose || (len(request.args) > 0 && (request.args[0] == "-v" || request.args[0] == "--verbose")) {
		out("\n")
		out("\t%s:\t%s\n", "home", "https://github.com/creativeprojects/resticprofile")
		out("\t%s:\t%s\n", "os", runtime.GOOS)
		out("\t%s:\t%s\n", "arch", runtime.GOARCH)
		if goarm > 0 {
			out("\t%s:\tv%d\n", "arm", goarm)
		}
		out("\t%s:\t%s\n", "version", version)
		out("\t%s:\t%s\n", "commit", commit)
		out("\t%s:\t%s\n", "compiled", date)
		out("\t%s:\t%s\n", "by", builtBy)
		out("\t%s:\t%s\n", "go version", runtime.Version())
		out("\n")
		out("\t%s:\n", "go modules")
		bi, _ := debug.ReadBuildInfo()
		for _, dep := range bi.Deps {
			out("\t\t%s\t%s\n", ansi.Cyan(dep.Path), dep.Version)
		}
		out("\n")
	}
	return nil
}

func displayProfilesCommand(output io.Writer, request commandRequest) error {
	displayProfiles(output, request.config, request.flags)
	displayGroups(output, request.config, request.flags)
	return nil
}

func displayProfiles(output io.Writer, configuration *config.Config, flags commandLineFlags) {
	out, closer := displayWriter(output)
	defer closer()

	profiles := configuration.GetProfiles()
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

func displayGroups(output io.Writer, configuration *config.Config, flags commandLineFlags) {
	out, closer := displayWriter(output)
	defer closer()

	groups := configuration.GetProfileGroups()
	if len(groups) == 0 {
		return
	}
	out("%s (name, profiles, description):\n", ansi.Bold("Groups available"))
	for name, groupList := range groups {
		out("\t%s:\t[%s]\t%s\n", name, ansi.Cyan(strings.Join(groupList.Profiles, ", ")), groupList.Description)
	}
	out("\n")
}
