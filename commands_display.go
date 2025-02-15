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
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

var (
	ansiBold   = color.New(color.Bold).SprintFunc()
	ansiCyan   = color.New(color.FgCyan).SprintFunc()
	ansiYellow = color.New(color.FgYellow).SprintFunc()
)

func displayWriter(output io.Writer, flags commandLineFlags) (out func(args ...any) io.Writer, closer func()) {
	if term.GetOutput() == output {
		output = term.GetColorableOutput()

		if width, _ := term.OsStdoutTerminalSize(); width > 10 {
			output = newLineLengthWriter(output, width)
		}
	}

	if flags.noAnsi {
		output = colorable.NewNonColorable(output)
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
		ansiBold("resticprofile"), profile, ansiBold(commandName),
	)
}

func displayOwnCommands(output io.Writer, ctx commandContext) {
	out, closer := displayWriter(output, ctx.flags)
	defer closer()

	for _, command := range ctx.ownCommands.commands {
		if command.hide {
			continue
		}

		out("\t%s\t%s\n", command.name, command.description)
	}
}

func displayOwnCommandHelp(output io.Writer, commandName string, ctx commandContext) {
	out, closer := displayWriter(output, ctx.flags)
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

func displayCommonUsageHelp(output io.Writer, ctx commandContext) {
	out, closer := displayWriter(output, ctx.flags)
	defer closer()

	out("resticprofile is a configuration profiles manager for backup profiles and ")
	out("is the missing link between a configuration file and restic backup\n\n")

	out("Usage:\n")
	out("\t%s [restic flags]\n", getCommonUsageHelpLine("restic-command", true))
	out("\t%s [command specific flags]\n", getCommonUsageHelpLine("resticprofile-command", true))
	out("\n")
	out(ansiBold("resticprofile flags:\n"))
	out(ctx.flags.usagesHelp)
	out("\n\n")
	out(ansiBold("resticprofile own commands:\n"))
	displayOwnCommands(out(), ctx)
	out("\n")

	out("%s at %s\n",
		ansiBold("Documentation available"),
		ansiBold(ansiCyan("https://creativeprojects.github.io/resticprofile/")))
	out("\n")
}

func displayResticHelp(output io.Writer, configuration *config.Config, flags commandLineFlags, command string) {
	out, closer := displayWriter(output, flags)
	defer closer()

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
		out("\nFlags applied by resticprofile (configuration \"%s\"):\n\n", ansiBold(configuration.GetConfigFile()))

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
				out("\tprofile \"%s\":", ansiBold(name))
				profile := profiles[name]
				cmdFlags := config.GetNonConfidentialArgs(profile, profile.GetCommandFlags(command))
				for _, flag := range cmdFlags.GetAll() {
					if strings.HasPrefix(flag, "-") {
						out("\n\t\t")
					}
					out("%s\t", ansiCyan(unescaper.Replace(flag)))
				}
				out("\n")
			}
		} else {
			out("none\n")
		}
	}
}

func displayHelpCommand(output io.Writer, ctx commandContext) error {
	flags := ctx.flags

	out, closer := displayWriter(output, ctx.flags)
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
		displayCommonUsageHelp(out("\n"), ctx)

	} else if ctx.ownCommands.Exists(*helpForCommand, true) || ctx.ownCommands.Exists(*helpForCommand, false) {
		displayOwnCommandHelp(out("\n"), *helpForCommand, ctx)

	} else {
		displayResticHelp(out(), ctx.config, flags, *helpForCommand)
	}

	return nil
}

func displayVersion(output io.Writer, ctx commandContext) error {
	out, closer := displayWriter(output, ctx.flags)
	defer closer()

	out("resticprofile version %s commit %s\n", ansiBold(version), ansiYellow(commit))

	// allow for the general verbose flag, or specified after the command
	arguments := ctx.request.arguments
	if ctx.flags.verbose || (len(arguments) > 0 && (arguments[0] == "-v" || arguments[0] == "--verbose")) {
		out("\n")
		out("\t%s:\t%s\n", "home", "https://github.com/creativeprojects/resticprofile")
		out("\t%s:\t%s\n", "os", runtime.GOOS)
		out("\t%s:\t%s\n", "arch", runtime.GOARCH)
		out("\t%s:\t%s\n", "version", version)
		out("\t%s:\t%s\n", "commit", commit)
		out("\t%s:\t%s\n", "compiled", date)
		out("\t%s:\t%s\n", "by", builtBy)
		out("\t%s:\t%s\n", "go version", runtime.Version())
		out("\n")
		out("\t%s:\n", "go modules")
		bi, _ := debug.ReadBuildInfo()
		for _, dep := range bi.Deps {
			out("\t\t%s\t%s\n", ansiCyan(dep.Path), dep.Version)
		}
		out("\n")
	}
	return nil
}

func displayProfilesCommand(output io.Writer, ctx commandContext) error {
	displayProfiles(output, ctx.config, ctx.flags)
	displayGroups(output, ctx.config, ctx.flags)
	return nil
}

func displayProfiles(output io.Writer, configuration *config.Config, flags commandLineFlags) {
	out, closer := displayWriter(output, flags)
	defer closer()

	profiles := configuration.GetProfiles()
	keys := sortedProfileKeys(profiles)
	if len(profiles) == 0 {
		out("\nThere's no available profile in the configuration\n")
	} else {
		out("\n%s (name, sections, description):\n", ansiBold("Profiles available"))
		for _, name := range keys {
			sections := profiles[name].DefinedCommands()
			sort.Strings(sections)
			if len(sections) == 0 {
				out("\t%s:\t(n/a)\t%s\n", name, profiles[name].Description)
			} else {
				out("\t%s:\t(%s)\t%s\n", name, ansiCyan(strings.Join(sections, ", ")), profiles[name].Description)
			}
		}
	}
	out("\n")
}

func displayGroups(output io.Writer, configuration *config.Config, flags commandLineFlags) {
	out, closer := displayWriter(output, flags)
	defer closer()

	groups := configuration.GetProfileGroups()
	if len(groups) == 0 {
		return
	}
	out("%s (name, profiles, description):\n", ansiBold("Groups available"))
	for name, groupList := range groups {
		out("\t%s:\t[%s]\t%s\n", name, ansiCyan(strings.Join(groupList.Profiles, ", ")), groupList.Description)
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

// lineLengthWriter limits the max line length, adding line breaks ('\n') as needed.
// the writer detects the right most column (consecutive whitespace) and aligns content if possible.
type lineLengthWriter struct {
	writer                                            io.Writer
	tokens                                            []byte
	maxLineLength, lastWhite, breakLength, lineLength int
	ansiLength, lastWhiteAnsiLength                   int
}

func newLineLengthWriter(writer io.Writer, maxLineLength int) *lineLengthWriter {
	return &lineLengthWriter{writer: writer, maxLineLength: maxLineLength}
}

func (l *lineLengthWriter) Write(p []byte) (n int, err error) {
	written := 0
	inAnsi := false
	offset := l.lineLength
	lineLength := func() int { return l.lineLength - l.ansiLength }

	if len(l.tokens) == 0 {
		l.tokens = []byte{' ', '\n'}
	}

	for i := 0; i < len(p); i++ {
		l.lineLength++
		ws := p[i] == l.tokens[0] // ' '
		br := p[i] == l.tokens[1] // '\n'

		// don't count ansi control sequences
		if inAnsi = inAnsi || p[i] == 0x1b; inAnsi {
			inAnsi = p[i] != 'm'
			l.ansiLength++
			continue
		}

		if !br && lineLength() > l.maxLineLength && l.lastWhite-offset > 0 {
			lastWhiteIndex := l.lastWhite - offset - 1
			remainder := i - lastWhiteIndex

			if written, err = l.writer.Write(p[:lastWhiteIndex]); err == nil {
				p = p[lastWhiteIndex+1:]
				i = remainder - 1
				n += written + 1

				_, _ = l.writer.Write(l.tokens[1:]) // write break (instead of WS at lastWhiteIndex)
				for j := 0; j < l.breakLength; j++ {
					_, _ = l.writer.Write(l.tokens[0:1]) // fill spaces for alignment
				}

				l.lineLength = l.breakLength + remainder
				l.lastWhite = l.breakLength
				offset = l.breakLength

				l.ansiLength -= l.lastWhiteAnsiLength
				l.lastWhiteAnsiLength = 0
			} else {
				return
			}
		}

		if ws {
			if l.lastWhite == l.lineLength-1 && lineLength() < l.maxLineLength*2/3 {
				l.breakLength = lineLength()
			}
			l.lastWhite = l.lineLength
			l.lastWhiteAnsiLength = l.ansiLength

		} else if br {
			if written, err = l.writer.Write(p[:i+1]); err == nil {
				p = p[i+1:]
				i = -1
				n += written

				l.lineLength = 0
				l.lastWhite = 0
				l.breakLength = 0
				offset = 0

				l.ansiLength = 0
				l.lastWhiteAnsiLength = 0
			} else {
				return
			}
		}
	}

	// write remainder
	if written, err = l.writer.Write(p); err == nil {
		n += written
	}
	return
}
