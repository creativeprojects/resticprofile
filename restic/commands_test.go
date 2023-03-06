package restic

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func descriptionIs(expected string) func(t *testing.T, cmd CommandIf, err error) {
	return func(t *testing.T, cmd CommandIf, err error) {
		assert.NoError(t, err)
		assert.Equal(t, expected, cmd.GetDescription())
	}
}

func optionNotExists(name string) func(t *testing.T, cmd CommandIf, err error) {
	return func(t *testing.T, cmd CommandIf, err error) {
		_, found := cmd.Lookup(name)
		assert.False(t, found)
		assert.False(t, slices.ContainsFunc(cmd.GetOptions(), func(option Option) bool { return option.Name == name }))
	}
}

func optionIs(name, description, def string, once bool) func(t *testing.T, cmd CommandIf, err error) {
	return func(t *testing.T, cmd CommandIf, err error) {
		assert.NoError(t, err)
		for _, n := range strings.Split(name, ",") {
			option, found := cmd.Lookup(n)
			require.Truef(t, found, "name %s not found", n)
			assert.Contains(t, cmd.GetOptions(), option, "lookup found an option that is not in GetOptions")
			if len(n) == 1 {
				assert.Equal(t, n, option.Alias, "alias for option: %s", name)
			} else {
				assert.Equal(t, n, option.Name, "name for option: %s", name)
			}
			assert.Equal(t, description, option.Description, "description for option: %s", name)
			assert.Equal(t, def, option.Default, "default for option: %s", name)
			assert.Equal(t, once, option.Once, "once for option: %s", name)
		}
	}
}

func all(checks ...func(t *testing.T, cmd CommandIf, err error)) func(t *testing.T, cmd CommandIf, err error) {
	return func(t *testing.T, cmd CommandIf, err error) {
		require.NotEmpty(t, checks)
		for _, check := range checks {
			check(t, cmd, err)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name, source string
		validation   func(t *testing.T, cmd CommandIf, err error)
	}{
		{
			name: "parse description",
			source: `.nh
.TH head

.SH SYNOPSIS
.PP
\fBrestic init [flags]\fP

.SH ANOTHER SECTION

.SH DESCRIPTION
.PP
The "init" command initializes a new repository.

.SH NEXT SECTION
.PP
Next paragraph
`,
			validation: descriptionIs("The \"init\" command initializes a new repository."),
		},

		{
			name: "parse multiline description",
			source: `
.SH DESCRIPTION
.PP
	First Line
	Second Line`,
			validation: descriptionIs("First Line\n\tSecond Line"),
		},

		{
			name: "merge description sections",
			source: `
.SH DESCRIPTION
.PP
First Line

.SH DESCRIPTION
.PP
Second Line`,
			validation: descriptionIs("First Line\n\nSecond Line"),
		},

		{
			name: "merge description paragraphs",
			source: `
.SH DESCRIPTION
.PP
First Line

.PP
Second Line`,
			validation: descriptionIs("First Line\n\nSecond Line"),
		},

		{
			name: "parse option",
			source: `
.SH DESCRIPTION
.PP
Desc 1

.SH OPTIONS
.PP
\fB--copy-chunker-params\fP[=false]
	copy chunker parameters from the secondary repository (useful with the copy command)

.SH DESCRIPTION
.PP
Desc 2
`,
			validation: all(
				descriptionIs("Desc 1\n\nDesc 2"),
				optionIs(
					"copy-chunker-params",
					"copy chunker parameters from the secondary repository (useful with the copy command)",
					"false",
					true,
				),
			),
		},

		{
			name: "parse multiline option",
			source: `
.SH OPTIONS
.PP
Ignored line

\fB--name,-n\fP=default-value
	First line.
	Second line.
Third line.

Fourth line.

.PP
Ignored line`,
			validation: optionIs(
				"n,name",
				"First line. Second line. Third line.  Fourth line.",
				"default-value",
				true,
			),
		},

		{
			name: "parse multiple options",
			source: `
.SH OPTIONS
.PP
\fB-o\fP, \fB--option\fP=[]
	set extended option (\fB\fCkey=value\fR, can be specified multiple times)

.PP
\fB--from-repo\fP=""
	source \fB\fCrepository\fR to copy chunker parameters from (default: $RESTIC_FROM_REPOSITORY)

.PP
\fB-h\fP, \fB--help\fP[=false]
	help for init

.PP
\fB--repository-version\fP="stable"
	repository format version to use, allowed values are a format version, 'latest' and 'stable'

.SH OPTIONS INHERITED FROM PARENT COMMANDS
.PP
\fB--cacert\fP=[]
	\fB\fCfile\fR to load root certificates from (default: use system certificates)
`,
			validation: all(
				optionIs(
					"o,option",
					"set extended option (key=value, can be specified multiple times)",
					"",
					false,
				),
				optionIs(
					"from-repo",
					"source repository to copy chunker parameters from (default: $RESTIC_FROM_REPOSITORY)",
					`""`,
					true,
				),
				optionIs(
					"h,help",
					"help for init",
					"false",
					true,
				),
				optionIs(
					"repository-version",
					"repository format version to use, allowed values are a format version, 'latest' and 'stable'",
					`"stable"`,
					true,
				),
				optionNotExists("cacert"),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd, err := parseStream(strings.NewReader(test.source), "")
			assert.Equal(t, DefaultCommand, cmd.Name)
			test.validation(t, cmd, err)
		})
	}
}

//go:embed fixtures
var fixtures embed.FS

func parseFixtures(t *testing.T) map[string]*command {
	cmds := map[string]*command{}
	for i, version := range []string{"0.9", "0.10", "0.14"} {
		manualDir, err := fs.Sub(fixtures, path.Join("fixtures", version))
		require.NoError(t, err)
		require.NotNil(t, manualDir)
		require.NoError(t, parseCommandsFromManPagesInto(manualDir, version, i == 0, cmds))
	}

	require.Contains(t, cmds, "init")
	require.Contains(t, cmds, "copy")
	return cmds
}

func TestParseFromFS(t *testing.T) {
	defer LoadEmbeddedCommands()

	version := "0.14"
	manualDir, err := fs.Sub(fixtures, path.Join("fixtures", version))
	assert.NoError(t, err)

	ClearCommands()
	assert.Empty(t, CommandNames())
	assert.NoError(t, ParseCommandsFromManPages(manualDir, version, true))
	assert.NotEmpty(t, CommandNames())
}

func TestParseVersionsFromFS(t *testing.T) {
	cmds := parseFixtures(t)

	copyCmd := cmds["copy"]
	init := cmds["init"]
	init09 := &commandAtVersion{command: *init, actualVersion: tryParseVersion("0.9")}
	init10 := &commandAtVersion{command: *init, actualVersion: tryParseVersion("0.10")}
	init14 := &commandAtVersion{command: *init, actualVersion: tryParseVersion("0.14")}

	t.Run("parsing succeeded", func(t *testing.T) {
		all(
			descriptionIs("The \"init\" command initializes a new repository."),
			optionIs(
				"from-key-hint",
				"key ID of key to try decrypting the source repository first (default: $RESTIC_FROM_KEY_HINT)",
				`""`,
				true),
			optionIs(
				"verbose",
				"be verbose (specify multiple times or a level using --verbose=n, max level/times is 3)",
				`0`,
				true),
		)(t, init, nil)
	})

	t.Run("verbose flag was removed in 0.14", func(t *testing.T) {
		// Existing in 0.9 (only in fixtures, not in real restic)
		verboseExists := optionIs(
			"verbose",
			"be verbose (specify multiple times or a level using --verbose=n, max level/times is 3)",
			`0`,
			true)
		verboseExists(t, init09, nil)
		verboseExists(t, init10, nil)
		// Removed in 0.14
		optionNotExists("verbose")(t, init14, nil)
		// Existing in 0.14 when includeRemoved is true
		init14WithRemoved := *init14
		init14WithRemoved.includeRemoved = true
		verboseExists(t, &init14WithRemoved, nil)
	})

	t.Run("test flag was removed in 0.10", func(t *testing.T) {
		// Existing in 0.9 (only in fixtures, not in real restic)
		testExists := optionIs(
			"test",
			"unused parameter for unit test",
			`""`,
			true)
		testExists(t, init09, nil)
		// Removed in 0.10
		optionNotExists("test")(t, init10, nil)
		optionNotExists("test")(t, init14, nil)
		// Existing in 0.14 when includeRemoved is true
		init14WithRemoved := *init14
		init14WithRemoved.includeRemoved = true
		testExists(t, &init14WithRemoved, nil)
	})

	t.Run("from-key-hint flag was added", func(t *testing.T) {
		// Unknown in 0.9
		optionNotExists("from-key-hint")(t, init09, nil)
		// Added in 0.14
		optionIs(
			"from-key-hint",
			"key ID of key to try decrypting the source repository first (default: $RESTIC_FROM_KEY_HINT)",
			`""`,
			true)(t, init14, nil)
	})

	t.Run("copy is not in all versions", func(t *testing.T) {
		// fixtures just cover 0.9 and 0.14
		excluded := []string{"0.13", "0.13.99", "0.12", "0.11", "0.9", "0.0"}
		included := []string{"0.14", "0.14.0", "0.14.1", "0.15", "1.0", "20.0"}
		for _, v := range included {
			assert.True(t, copyCmd.ContainedInVersion(v))
		}
		for _, v := range excluded {
			assert.False(t, copyCmd.ContainedInVersion(v))
		}
	})

	t.Run("return versions with changes", func(t *testing.T) {
		// fixtures just cover 0.9 (base), 0.10 and 0.14
		expected := []string{"0.14", "0.10", "0.9"}
		assert.Equal(t, expected, knownVersionsFrom(cmds))
	})
}

func TestLoadAndSave(t *testing.T) {
	defer LoadEmbeddedCommands()

	t.Run("store", func(t *testing.T) {
		file := path.Join(t.TempDir(), "commands.json")
		LoadEmbeddedCommands()

		contents := make([][]byte, 2)
		for i := 0; i < 2; i++ {
			err := StoreCommands(file)
			assert.NoError(t, err)
			contents[i], err = os.ReadFile(file)
			assert.NoError(t, err)
			assert.NoError(t, os.Truncate(file, 0))
		}

		// Content is stable
		assert.Equal(t, string(contents[0]), string(contents[1]))

		// Content is sorted
		names, idx := CommandNames(), 0
		sort.Strings(names)
		for _, name := range names {
			index := bytes.Index(contents[0], []byte(fmt.Sprintf(`%s    "Name": "%s"`, "\n", name)))
			assert.Greater(t, index, idx, "index of %s greater %d", name, idx)
			idx = index
		}
	})

	t.Run("load-no-replace-on-failure", func(t *testing.T) {
		file := path.Join(t.TempDir(), "commands.json")
		LoadEmbeddedCommands()
		names := CommandNames()
		assert.ErrorContains(t, LoadCommands(file), fmt.Sprintf("open %s:", file))
		// ensure load failure did not replace existing commands
		assert.Equal(t, names, CommandNames())
	})

	t.Run("load-replace", func(t *testing.T) {
		file := path.Join(t.TempDir(), "commands.json")
		cmds := parseFixtures(t)
		assert.NoError(t, storeCommands(cmds, file))

		LoadEmbeddedCommands()
		names := CommandNames()
		assert.NoError(t, LoadCommands(file))

		assert.NotEqual(t, names, CommandNames())
		assert.Equal(t, commandNamesForVersion(cmds, AnyVersion), CommandNames())
	})

	t.Run("load", func(t *testing.T) {
		cmds := parseFixtures(t)
		file := path.Join(t.TempDir(), "commands.json")
		assert.NoError(t, storeCommands(cmds, file))

		loaded, err := loadCommands(file)
		assert.NoError(t, err)
		assert.Equal(t, cmds, loaded)
	})
}

func TestBuiltInCommandsTable(t *testing.T) {
	LoadEmbeddedCommands()

	expectedCommands := []string{
		"backup", "cache", "cat", "check",
		"copy", "diff", "dump", "find",
		"forget", "generate", "init", "key",
		"list", "ls", "migrate", "mount",
		"prune", "rebuild-index", "recover", "restore",
		"self-update", "snapshots", "stats", "tag",
		"unlock", "version",
	}

	t.Run("available commands", func(t *testing.T) {
		commands09 := collect.All(expectedCommands, func(cmd string) bool {
			return cmd != "copy"
		})

		assert.Subset(t, CommandNames(), expectedCommands)
		assert.Equal(t, expectedCommands, CommandNamesForVersion("0.14"))
		assert.Equal(t, expectedCommands, CommandNamesForVersion("0.10"))
		assert.Equal(t, commands09, CommandNamesForVersion("0.9"))
		assert.Equal(t, commands09, CommandNamesForVersion("0.0"))
	})

	t.Run("get commands", func(t *testing.T) {
		for _, name := range expectedCommands {
			cmd, found := GetCommand(name)
			assert.NotNil(t, cmd)
			assert.True(t, found)
		}
	})

	t.Run("no copy in 0.9", func(t *testing.T) {
		cmd, found := GetCommandForVersion("copy", "0.9", false)
		assert.Nil(t, cmd)
		assert.False(t, found)
	})

	t.Run("available default options", func(t *testing.T) {
		expectedOptions := []string{
			"cacert",
			"cache-dir",
			"cleanup-cache",
			"compression",
			"help",
			"insecure-tls",
			"json",
			"key-hint",
			"limit-download",
			"limit-upload",
			"no-cache",
			"no-lock",
			"option",
			"pack-size",
			"password-command",
			"password-file",
			"quiet",
			"repo",
			"repository-file",
			"tls-client-cert",
			"verbose",
		}
		for _, option := range GetDefaultOptions() {
			assert.Contains(t, expectedOptions, option.Name)
			assert.True(t, option.AvailableForOS())
		}
		assert.GreaterOrEqual(t, len(GetDefaultOptions()), len(expectedOptions))
	})

	t.Run("windows specific option", func(t *testing.T) {
		cmd, _ := GetCommandForVersion("backup", "0.12", false)
		require.NotNil(t, cmd)
		option, found := cmd.Lookup("use-fs-snapshot")
		require.True(t, found)

		assert.False(t, option.AvailableInOS("linux"))
		assert.False(t, option.AvailableInOS("darwin"))
		assert.True(t, option.AvailableInOS("windows"))
	})
}
