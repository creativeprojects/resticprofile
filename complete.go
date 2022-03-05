package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/spf13/pflag"
)

const (
	RequestFileCompletion   = "__complete_file"
	RequestResticCompletion = "__complete_restic"
	ProfileCommandDelimiter = "."
	CursorPositionDelimiter = ":"
	CursorPositionPrefix    = "__POS" + CursorPositionDelimiter
)

// Completer provides context aware shell completions for the current commandline argument(s)
type Completer struct {
	flags                 *pflag.FlagSet
	flagsInArgs           []*pflag.Flag
	ownCommands           []ownCommand
	profiles              []string
	enableProfilePrefixes bool
}

func (c *Completer) init(args []string) {
	var initArgs []string
	nameFlagFound := false
	nameFlagMatcher := regexp.MustCompile("^-{1,2}(n|name)(=.*|$)")

	for _, arg := range args {
		if nameFlagMatcher.MatchString(arg) {
			nameFlagFound = true
		}
		if !strings.HasPrefix(arg, CursorPositionPrefix) {
			initArgs = append(initArgs, arg)
		}
	}

	c.flags, _, _ = loadFlags(initArgs)
	c.flagsInArgs = nil
	c.ownCommands = getOwnCommands()
	c.profiles = nil
	c.enableProfilePrefixes = !nameFlagFound
}

func (c *Completer) formatFlag(flag *pflag.Flag, shorthand bool) string {
	if shorthand {
		return fmt.Sprintf("-%s", flag.Shorthand)
	} else {
		return fmt.Sprintf("--%s", flag.Name)
	}
}

func (c *Completer) completeFlagSet(word string) (completions []string) {
	c.flags.SortFlags = true
	c.flags.VisitAll(func(flag *pflag.Flag) {
		// Skip already specified flags
		for _, f := range c.flagsInArgs {
			if f.Name == flag.Name {
				return
			}
		}

		if flag.Hidden {
			return
		}

		if len(word) == 0 {
			completions = append(completions, c.formatFlag(flag, false))
		} else {
			names := []string{
				c.formatFlag(flag, true),
				c.formatFlag(flag, false),
			}
			completions = c.appendMatches(completions, word, names...)
		}
	})

	return
}

func (c *Completer) completeFlagSetValue(flag *pflag.Flag, word string) (completions []string) {
	var list []string

	switch flag.Name {
	case "name":
		list = c.listProfileNames()

	case "format":
		list = []string{"toml", "json", "yaml", "hcl"}

	case "theme":
		list = []string{"dark", "light", "none"}

	case "config":
		fallthrough
	case "log":
		completions = []string{RequestFileCompletion}
	}

	completions = c.appendMatches(completions, word, list...)
	return
}

func (c *Completer) listProfileNames() (list []string) {
	if c.profiles == nil {
		filename := ""
		format := ""
		if configFlag := c.flags.Lookup("config"); configFlag != nil {
			filename = configFlag.Value.String()
		}
		if formatFlag := c.flags.Lookup("format"); formatFlag != nil {
			format = formatFlag.Value.String()
		}

		if file, err := filesearch.FindConfigurationFile(filename); err == nil {
			if conf, err := config.LoadFile(file, format); err == nil {
				list = append(list, conf.GetProfileNames()...)
				for name, _ := range conf.GetProfileGroups() {
					list = append(list, name)
				}
			} else {
				clog.Debug(err)
			}
		} else {
			clog.Debug(err)
		}

		if list == nil {
			list = make([]string, 0)
		}
		sort.Strings(list)
		c.profiles = list
	} else {
		list = c.profiles
	}

	return
}

func (c *Completer) completeProfileNamePrefixes(word string) (completions []string) {
	if c.enableProfilePrefixes {
		for _, profile := range c.listProfileNames() {
			completions = c.appendMatches(completions, word, profile+ProfileCommandDelimiter)
		}
	}
	return
}

func (c *Completer) formatOwnCommand(command ownCommand) string {
	return command.name
}

func (c *Completer) completeOwnCommands(word string) (completions []string) {
	for _, command := range c.ownCommands {
		if command.hide || command.hideInCompletion {
			continue
		}
		completions = c.appendMatches(completions, word, c.formatOwnCommand(command))
	}

	if completions != nil {
		sort.Strings(completions)
	}
	return
}

func (c *Completer) completeOwnCommandFlags(name, word string) (completions []string) {
	commandFlagsInArgs := c.flags.Args()
	sort.Strings(commandFlagsInArgs)

	for _, command := range c.ownCommands {
		if command.name == name {
			for names, _ := range command.flags { // e.g. "-q, --quiet"
				var flagNames []string

				for _, flag := range strings.Split(names, ",") {
					flag = strings.TrimSpace(flag)
					flagNames = append(flagNames, flag)

					// Remove this flag if already specified
					if c.sortedListContains(commandFlagsInArgs, flag) {
						flagNames = nil
						break
					}
				}

				if flagNames != nil {
					completions = c.appendMatches(completions, word, flagNames...)
				}
			}

			if completions != nil {
				sort.Strings(completions)
			}
		}
	}

	return
}

func (c *Completer) appendMatches(completions []string, word string, list ...string) []string {
	for _, item := range list {
		if strings.HasPrefix(item, word) && item != "--" && len(item) > 1 {
			completions = append(completions, item)
		}
	}
	return completions
}

func (c *Completer) sortedListContains(list []string, word string) bool {
	// expects list to be sorted in ascending order (will fail if not)
	index := sort.SearchStrings(list, word)
	return index < len(list) && list[index] == word
}

func (c *Completer) toCompletionsWithProfilePrefix(completions []string, profile string) (result []string) {
	if c.enableProfilePrefixes && c.sortedListContains(c.listProfileNames(), profile) {
		prefix := profile + ProfileCommandDelimiter

		for _, item := range completions {
			supportsPrefix := item != prefix &&
				!strings.HasPrefix(item, "-") &&
				!strings.HasSuffix(item, ProfileCommandDelimiter) &&
				item != RequestFileCompletion &&
				!c.isOwnCommand(item, false)

			if supportsPrefix {
				result = append(result, prefix+item)
			}
		}
	} else {
		result = completions
	}

	return
}

func (c *Completer) isOwnCommand(command string, configurationLoaded bool) bool {
	for _, commandDef := range c.ownCommands {
		if commandDef.name == command && commandDef.needConfiguration == configurationLoaded {
			return true
		}
	}
	return false
}

// Complete returns shell completions for the specified commandline args.
// By default, completions are generated for the last element in args.
// To get completions for a specific element use "__POS:{FOLLOWING-ARG-INDEX}" in args.
func (c *Completer) Complete(args []string) (completions []string) {
	args = append([]string{}, args...)
	c.init(args)

	lookupFlag := func(word string) (flag *pflag.Flag) {
		if strings.HasPrefix(word, "-") {
			flagName := strings.Trim(word, "-")
			flag = c.flags.Lookup(flagName)
			if flag == nil && len(flagName) == 1 && !strings.HasPrefix(word, "--") {
				flag = c.flags.ShorthandLookup(flagName)
			}
		}
		return
	}

	flagRequiresValue := func(flag *pflag.Flag) bool {
		return flag.Value.Type() != "bool"
	}

	isAnyOwnCommand := func(command string) bool {
		return len(command) > 0 && (c.isOwnCommand(command, false) || c.isOwnCommand(command, true))
	}

	// Parse cursor position (specified as "__POS:{FOLLOWING-ARG-INDEX}")
	for i, length := 0, len(args); i < length; i++ {
		if strings.HasPrefix(args[i], CursorPositionPrefix) {
			if pos, err := strconv.Atoi(strings.Split(args[i], CursorPositionDelimiter)[1]); err == nil && pos >= 0 {
				if pos < length-i {
					length = i + pos + 1
					args = args[0:length]
				} else {
					args = append(args, "")
					length++
				}
			} else {
				clog.Debugf("failed on arg \"%s\" ; cause %v", args[i], err)
			}
			args = append(args[0:i], args[i+1:]...)[0 : length-1]
			break
		}
	}

	// Parsing CLI args and setting the completion context
	var currentFlag *pflag.Flag
	inFlagSet := true
	atCommand := false
	word, commandName, profileName := "", "", ""

	for _, arg := range args {
		currentFlag = nil
		atCommand = false

		// Unknown command: Delegate to restic completion
		if !inFlagSet && !isAnyOwnCommand(commandName) {
			completions = []string{RequestResticCompletion}
			break
		}

		// Handle own flags and commands
		if inFlagSet {
			if strings.HasPrefix(arg, "-") {
				if flag := lookupFlag(arg); flag != nil {
					c.flagsInArgs = append(c.flagsInArgs, flag)
				}
			} else {
				inFlagSet = false

				if currentFlag = lookupFlag(word); currentFlag != nil {
					inFlagSet = flagRequiresValue(currentFlag)
				}

				if !inFlagSet {
					atCommand = true
					commandName = arg
					if pos := strings.LastIndex(commandName, ProfileCommandDelimiter); pos > -1 {
						profileName, commandName = commandName[0:pos], commandName[pos+1:]
						word = commandName
						continue
					}
				}
			}
		}

		word = arg
	}

	// Build completions
	if completions == nil {
		if inFlagSet {
			// Test for exact flag match in current input
			if f := lookupFlag(word); f != nil {
				if flagRequiresValue(f) {
					currentFlag = f
				}
				word = ""
			}

			if currentFlag == nil {
				completions = c.completeFlagSet(word)
				completions = append(completions, c.completeOwnCommands(word)...)
				completions = append(completions, c.completeProfileNamePrefixes(word)...)
			} else if len(c.flagsInArgs) > 0 {
				completions = c.completeFlagSetValue(currentFlag, word)
			}
		} else if len(commandName) > 0 && isAnyOwnCommand(commandName) {
			completions = c.completeOwnCommandFlags(commandName, word)
		} else {
			completions = c.completeOwnCommands(word)
			completions = append(completions, c.completeProfileNamePrefixes(word)...)
			completions = append(completions, RequestResticCompletion)
		}
	}

	// Remove direct matches to avoid duplications
	if len(completions) == 1 && completions[0] == word {
		completions = nil
	}

	// Add command profile prefix when a valid profile name is specified
	if len(profileName) > 0 && atCommand {
		completions = c.toCompletionsWithProfilePrefix(completions, profileName)
	}

	return
}
