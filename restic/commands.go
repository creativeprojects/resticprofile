package restic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/util/collect"
)

const (
	AnyVersion     = ""
	BaseVersion    = "0.9"
	DefaultCommand = "__default"
)

// Option provides meta information for a restic command option
type Option struct {
	Name, Alias, Default, Description string
	Once                              bool
	OnlyInOS                          []string `json:",omitempty"`
	FromVersion, RemovedInVersion     string
}

func (o *Option) GetFromVersion() string      { return o.FromVersion }
func (o *Option) GetRemovedInVersion() string { return o.RemovedInVersion }
func (o *Option) ContainedInVersion(version string) bool {
	return includedInVersion(o, false, tryParseVersion(version))
}

// AvailableForOS returns true if the option is available in the current runtime.GOOS
func (o *Option) AvailableForOS() bool { return o.AvailableInOS(runtime.GOOS) }

// AvailableInOS returns true if the option is available in the specified goos
func (o *Option) AvailableInOS(goos string) bool {
	return len(o.OnlyInOS) == 0 || slices.Contains(o.OnlyInOS, goos)
}

// Command provides meta information for a restic command
type command struct {
	Name, Description             string
	FromVersion, RemovedInVersion string
	Options                       []Option
}

func (c *command) GetName() string             { return c.Name }
func (c *command) GetDescription() string      { return c.Description }
func (c *command) GetFromVersion() string      { return c.FromVersion }
func (c *command) GetRemovedInVersion() string { return c.RemovedInVersion }
func (c *command) ContainedInVersion(version string) bool {
	return includedInVersion(c, false, tryParseVersion(version))
}

func (c *command) GetOptions() []Option {
	return append(make([]Option, 0), c.Options...)
}

func (c *command) Lookup(name string) (option Option, found bool) {
	option, found = c.lookupOptions(name, c.Options)
	if !found && c.Name != DefaultCommand {
		if defCommand, exists := commands[DefaultCommand]; exists {
			option, found = c.lookupOptions(name, defCommand.Options)
		}
	}
	return
}

func (c *command) lookupOptions(name string, options []Option) (option Option, found bool) {
	for _, opt := range options {
		if opt.Name == name || (len(opt.Alias) > 0 && opt.Alias == name) {
			option = opt
			found = true
			return
		}
	}
	return
}

func (c *command) sortOptions() {
	if len(c.Options) > 0 {
		sort.Slice(c.Options, func(i, j int) bool {
			return c.Options[i].Name < c.Options[j].Name
		})
	}
}

type commandAtVersion struct {
	command

	includeRemoved bool
	actualVersion  *semver.Version
}

func (v *commandAtVersion) GetOptions() (opts []Option) {
	for _, o := range v.Options {
		if includedInVersion(&o, v.includeRemoved, v.actualVersion) {
			opts = append(opts, o)
		} else {
			clog.Tracef("skipping %s option %s (from: %s, removed-in: %s, actual: %s)",
				v.Name, o.Name, o.FromVersion, o.RemovedInVersion, v.actualVersion)
		}
	}
	return
}

func (v *commandAtVersion) Lookup(name string) (option Option, found bool) {
	option, found = v.command.Lookup(name)
	if found {
		found = includedInVersion(&option, v.includeRemoved, v.actualVersion)
	}
	return
}

// CommandIf provides access to shared Command instances
type CommandIf interface {
	Versioned
	// GetName provides the command name (e.g. "backup")
	GetName() string
	// GetDescription provides the long description of the command
	GetDescription() string
	// GetOptions returns a list of all valid options (excluding GetDefaultOptions)
	GetOptions() []Option
	// Lookup returns a named option if available
	Lookup(name string) (option Option, found bool)
}

// Versioned indicates that the item may be available only for certain restic versions
type Versioned interface {
	// GetFromVersion returns the version when the item was supported initially
	GetFromVersion() string
	// GetRemovedInVersion returns the version when the item was no longer supported
	GetRemovedInVersion() string
	// ContainedInVersion is true the item is contained in the specified version
	ContainedInVersion(version string) bool
}

var (
	commands                map[string]*command
	latestVersionInCommands = AnyVersion

	manSectionStart   = regexp.MustCompile(`^\.SH ([A-Z 0-9]+)$`)
	manParagraphStart = regexp.MustCompile(`^\.PP$`)
	manBulletPoint    = regexp.MustCompile(`^\.IP \\\(bu \d+$`)
	manEscapeSequence = regexp.MustCompile(`(\\f[A-Z]|\\&|\\)`)
	diffCode          = regexp.MustCompile(`^([+-UMT?])  `)
	lineCleanup       = regexp.MustCompile("([`]{2,})")
)

func parseStream(input io.Reader, commandName string) (cmd *command, err error) {
	commandName = strings.Trim(commandName, "-. \t\n\r")
	if len(commandName) == 0 {
		commandName = DefaultCommand
	}

	bulletPoint := false

	cmd = &command{Name: commandName}

	var option *Option
	addOption := func() {
		if option != nil {
			option.Description = strings.TrimSpace(option.Description)
			cmd.Options = append(cmd.Options, *option)
		}
		option = nil
	}

	scanner := bufio.NewScanner(input)
	section := ""

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ".") {
			if m := manSectionStart.FindStringSubmatch(line); m != nil {
				section = m[1]
			} else if manParagraphStart.MatchString(line) {
				addOption()
			} else if manBulletPoint.MatchString(line) {
				bulletPoint = true
			} else {
				bulletPoint = false
			}
		} else {
			line = manEscapeSequence.ReplaceAllString(line, "")
			line = lineCleanup.ReplaceAllString(line, "")

			switch section {
			case "DESCRIPTION":
				if len(cmd.Description) > 0 {
					cmd.Description += "\n"
				}
				if bulletPoint {
					cmd.Description += "* "
					bulletPoint = false
					line = diffCode.ReplaceAllString(line, "`$1` ")
				}
				cmd.Description += line
			case "OPTIONS":
				if option == nil {
					if !strings.Contains(line, "--") {
						continue
					}
					option = &Option{Once: !strings.Contains(line, "=[]")}
					paramAndDefault := strings.Split(line, "=")
					for _, param := range strings.Split(paramAndDefault[0], ",") {
						param = strings.TrimSpace(param)
						if strings.HasPrefix(param, "--") {
							if len(option.Name) == 0 {
								option.Name = strings.Trim(param, "-[]")
							}
						} else if strings.HasPrefix(param, "-") {
							option.Alias = strings.Trim(param, "-[]")
						}
					}
					if len(paramAndDefault) > 0 {
						option.Default = strings.Trim(paramAndDefault[1], "[]")
					}
				} else {
					if len(option.Description) > 0 {
						option.Description += " "
					}
					option.Description += strings.TrimSpace(line)
				}
			}
		}
	}
	addOption()
	cmd.Description = strings.TrimSpace(cmd.Description)
	err = scanner.Err()
	return
}

func parseFile(manualDir fs.FS, filename, commandName string) (*command, error) {
	if file, err := manualDir.Open(filename); err == nil {
		defer file.Close()
		return parseStream(file, commandName)
	} else {
		return nil, err
	}
}

func tryParseVersion(version string) *semver.Version {
	if v, err := semver.NewVersion(version); err != nil {
		clog.Tracef("failed parsing restic version %q: %s", version, err.Error())
		return nil
	} else {
		return v
	}
}

// ParseCommandsFromManPages parses commands from manual pages located in the specified directory
// and adds them to the known commands. version indicates the restic version that generated the man pages and
// baseVersion indicates that the version is the min required base version and will not be used for filtering
func ParseCommandsFromManPages(manualDir fs.FS, version string, baseVersion bool) error {
	if commands == nil {
		commands = map[string]*command{}
	}
	return parseCommandsFromManPagesInto(manualDir, version, baseVersion, commands)
}

func parseCommandsFromManPagesInto(manualDir fs.FS, version string, baseVersion bool, commands map[string]*command) error {
	filePattern := regexp.MustCompile(`^restic(|-.+)\.1$`)
	newCommands := map[string]*command{}

	if files, err := fs.ReadDir(manualDir, "."); err == nil {
		for _, file := range files {
			if m := filePattern.FindStringSubmatch(file.Name()); m != nil && !file.IsDir() {
				var cmd *command
				if cmd, err = parseFile(manualDir, file.Name(), m[1]); err != nil {
					return err
				}

				if cmd != nil {
					cmd.FromVersion = version
					for i := range cmd.Options {
						cmd.Options[i].FromVersion = version
					}

					newCommands[cmd.Name] = cmd
				}
			}
		}
	} else {
		return err
	}

	base := ""
	if baseVersion {
		base = version
	}
	mergeCommandsInto(base, newCommands, commands)

	return nil
}

func mergeCommandsInto(baseVersion string, source, target map[string]*command) {
	sourceVersion := ""

	for _, sc := range source {
		sourceVersion = sc.FromVersion

		if tc := target[sc.Name]; tc == nil || len(tc.Options) == 0 {
			// Add new command
			sc.sortOptions()
			target[sc.Name] = sc

		} else {
			if len(tc.Description) < len(sc.Description) {
				tc.Description = sc.Description
			}

			// Merge options
			optionsSet := func(options []Option) (opts map[string]int) {
				opts = make(map[string]int)
				for i := range options {
					opts[options[i].Name] = i
				}
				return
			}

			// add new options
			targetOpts := optionsSet(tc.Options)
			for _, sopt := range sc.Options {
				if i, exists := targetOpts[sopt.Name]; exists {
					tc.Options[i].Alias = sopt.Alias
					tc.Options[i].Default = sopt.Default
					tc.Options[i].Description = sopt.Description
					tc.Options[i].Once = sopt.Once
				} else {
					tc.Options = append(tc.Options, sopt)
				}
			}

			// record removed options
			sourceOpts := optionsSet(sc.Options)
			for i, topt := range tc.Options {
				if _, exists := sourceOpts[topt.Name]; !exists && topt.RemovedInVersion == "" {
					tc.Options[i].RemovedInVersion = sourceVersion
				}
			}

			tc.sortOptions()
		}
	}

	// record removed commands (unlikely but for completeness)
	if len(sourceVersion) > 0 {
		for _, tc := range target {
			if source[tc.Name] == nil && tc.RemovedInVersion == "" {
				tc.RemovedInVersion = sourceVersion
			}
		}
	}

	// remove version from entries at base version or earlier
	if len(baseVersion) > 0 {
		if base := tryParseVersion(baseVersion); base != nil {
			for _, cmd := range target {
				if includedInVersion(cmd, true, base) {
					cmd.FromVersion = ""
				}
				for i := range cmd.Options {
					if includedInVersion(&cmd.Options[i], true, base) {
						cmd.Options[i].FromVersion = ""
					}
				}
			}
		}
	}
}

// StoreCommands stores current list of commands to the specified file
func StoreCommands(filename string) error { return storeCommands(commands, filename) }

// StoreCommands stores current list of commands to the specified file
func storeCommands(commands map[string]*command, filename string) error {
	names := commandNamesForVersion(commands, AnyVersion)
	sort.Strings(names)
	names = append(names, DefaultCommand)

	var list []*command
	for _, name := range names {
		if c := commands[name]; c != nil {
			c.sortOptions()
			list = append(list, c)
		}
	}

	if file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644); err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		return encoder.Encode(list)
	} else {
		return err
	}
}

// LoadCommands loads current list of commands from the specified file
func LoadCommands(filename string) (err error) {
	var cmds map[string]*command
	if cmds, err = loadCommands(filename); err == nil && len(cmds) > 0 {
		ClearCommands()
		commands = cmds
	}
	return
}

func loadCommands(filename string) (commands map[string]*command, err error) {
	var file io.ReadCloser
	if file, err = os.Open(filename); err == nil {
		defer file.Close()
		commands, err = loadCommandsFromReader(file)
	}
	return
}

func loadCommandsFromReader(reader io.Reader) (commands map[string]*command, err error) {
	var list []*command
	if err = json.NewDecoder(reader).Decode(&list); err != nil {
		return
	}

	if len(list) > 0 {
		commands = map[string]*command{}
		for _, cmd := range list {
			commands[cmd.Name] = cmd
		}
	}
	return
}

func loadCommandExtensionsFromReader(reader io.Reader) (extensions map[string][]Option, err error) {
	err = json.NewDecoder(reader).Decode(&extensions)
	return
}

func applyCommandExtensions(commands map[string]*command, extensions map[string][]Option) {
	for name, options := range extensions {
		if cmd, ok := commands[name]; ok {
			cmd.Options = append(cmd.Options, options...)
			cmd.sortOptions()
		}
	}
}

// ClearCommands removes all know restic commands
func ClearCommands() {
	latestVersionInCommands = AnyVersion
	commands = nil
}

func includedInVersion(item Versioned, includeRemoved bool, actual *semver.Version) bool {
	if actual != nil && item != nil {
		if from := item.GetFromVersion(); from != "" {
			if v := tryParseVersion(from); v == nil || v.GreaterThan(actual) {
				return false
			}
		}

		if removed := item.GetRemovedInVersion(); removed != "" && !includeRemoved {
			if v := tryParseVersion(removed); v == nil || actual.GreaterThan(v) || actual.Equal(v) {
				return false
			}
		}
	}
	return true
}

// KnownVersions returns all restic versions in descending order that applied changes to available commands or command flags
func KnownVersions() (versions []string) { return knownVersionsFrom(commands) }

func knownVersionsFrom(commands map[string]*command) (versions []string) {
	versions = []string{BaseVersion}
	for _, cmd := range commands {
		for _, option := range cmd.Options {
			if option.RemovedInVersion != "" {
				versions = append(versions, option.RemovedInVersion)
			}
			if option.FromVersion != "" {
				versions = append(versions, option.FromVersion)
			}
		}
	}

	fixedWidth := func(s string) string {
		return fmt.Sprintf("%6s", s)
	}
	slices.SortFunc(versions, func(a, b string) int {
		as := collect.From(strings.Split(a, "."), fixedWidth)
		bs := collect.From(strings.Split(b, "."), fixedWidth)
		return -slices.Compare(as, bs)
	})
	versions = slices.Compact(versions)
	return
}

func latestKnownVersion() string {
	if latestVersionInCommands == AnyVersion {
		latestVersionInCommands = KnownVersions()[0]
	}
	return latestVersionInCommands
}

// CommandNames returns the command names of the latest known restic version
func CommandNames() []string {
	return CommandNamesForVersion(VersionLatest)
}

// CommandNamesForVersion returns the names of all known restic commands for the specified restic version
func CommandNamesForVersion(version string) (names []string) {
	return commandNamesForVersion(commands, version)
}

func commandNamesForVersion(commands map[string]*command, version string) (names []string) {
	var actualVersion *semver.Version
	if version == VersionLatest {
		version = latestKnownVersion()
	}
	if version != "" {
		if actualVersion = tryParseVersion(version); actualVersion == nil {
			return
		}
	}

	for name, cmd := range commands {
		if name == DefaultCommand {
			continue
		}
		if !includedInVersion(cmd, false, actualVersion) {
			clog.Tracef("skipping restic command %s (actual-version: %s, from: %s, removed-in: %s)",
				name, actualVersion, cmd.FromVersion, cmd.RemovedInVersion)
			continue
		}
		names = append(names, name)
	}

	if names != nil {
		sort.Strings(names)
	}
	return
}

// GetDefaultOptions returns options that are valid for all commands for the latest known restic version
func GetDefaultOptions() []Option {
	return GetDefaultOptionsForVersion(VersionLatest, true)
}

// GetDefaultOptionsForVersion returns options that are valid for all commands for the specified restic version
func GetDefaultOptionsForVersion(version string, includeRemoved bool) []Option {
	if cmd, found := GetCommandForVersion(DefaultCommand, version, includeRemoved); found && cmd != nil {
		return cmd.GetOptions()
	}
	return nil
}

// GetCommand returns the named command for the latest known restic version
func GetCommand(name string) (CommandIf, bool) {
	return GetCommandForVersion(name, VersionLatest, true)
}

// GetCommandForVersion returns the named command for the specified restic version
func GetCommandForVersion(name, version string, includeRemoved bool) (command CommandIf, found bool) {
	cmd := commands[name]
	found = cmd != nil
	command = cmd

	if version == VersionLatest {
		version = latestKnownVersion()
	}
	if found && version != "" {
		actualVersion := tryParseVersion(version)
		if actualVersion != nil && includedInVersion(cmd, false, actualVersion) {
			command = &commandAtVersion{command: *cmd, includeRemoved: includeRemoved, actualVersion: actualVersion}
		} else {
			found = false
			command = nil
		}
	}
	return
}
