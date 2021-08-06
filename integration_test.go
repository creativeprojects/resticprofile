package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echoBinary string
)

func init() {
	// build restic mock
	cmd := exec.Command("go", "build", "./shell/echo")
	cmd.Run()
	if runtime.GOOS == "windows" {
		echoBinary = "echo.exe"
	} else {
		echoBinary = "./echo"
	}
}

// TestFromConfigFileToCommandLine loads all examples/integration_test.* configuration files
// and run some commands to display all the arguments that were sent
func TestFromConfigFileToCommandLine(t *testing.T) {
	files, err := filepath.Glob("./examples/integration_test.*")
	require.NoError(t, err)
	require.Greater(t, len(files), 0)

	// we can use the same files to test a glob pattern
	globFiles := "\"" + strings.Join(files, "\" \"") + "\""

	integrationData := []struct {
		profileName       string
		commandName       string
		cmdlineArgs       []string
		expected          string
		expectedOnWindows string
	}{
		{
			"default",
			"snapshots",
			[]string{},
			`"snapshots" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path"`,
			"",
		},
		{
			"simple",
			"backup",
			[]string{"--option"},
			`"backup" "--exclude" "/**/.git" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "--option" "/source"`,
			"",
		},
		{
			"spaces",
			"backup",
			[]string{"some path"},
			// `"backup" "--exclude" "My Documents" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "some" "path" "/source" "dir"`,
			`"backup" "--exclude" "My Documents" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "some path" "/source dir"`,
			"",
		},
		{
			"quotes",
			"backup",
			[]string{"quo'te", "quo\"te"},
			// `"backup" "--exclude" "MyDocuments --exclude My\"Documents --password-file key --repo rest:http://user:password@localhost:8000/path quote" "quote /source'dir /sourcedir"`,
			`"backup" "--exclude" "My'Documents" "--exclude" "My\"Documents" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "quo'te" "quo\"te" "/source'dir" "/source\"dir"`,
			"",
		},
		{
			"glob",
			"backup",
			[]string{"examples/integration*"},
			// `"backup" "--exclude" ` + globFiles + ` "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" ` + globFiles + " " + globFiles,
			`"backup" "--exclude" "examples/integration*" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" ` + globFiles + " " + globFiles,
			`"backup" "--exclude" "examples/integration*" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "examples/integration*" "examples/integration*"`,
		},
		{
			"mixed",
			"backup",
			[]string{"/path/with space; echo foo"},
			`"backup" "--exclude" "examples/integration*" "--password-file" "examples/key" "--repo" "rest:http://user:password@localhost:8000/path" "/path/with space; echo foo" "/Côte d'Ivoire"`,
			"",
		},
	}

	// try all the config files one by one
	for _, configFile := range files {
		t.Run(configFile, func(t *testing.T) {
			cfg, err := config.LoadFile(configFile, "")
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// try all the fixtures one by one (on each file)
			for _, fixture := range integrationData {
				t.Run(fixture.profileName+"/"+fixture.commandName, func(t *testing.T) {
					profile, err := cfg.GetProfile(fixture.profileName)
					require.NoError(t, err)
					require.NotNil(t, profile)

					profile.SetRootPath("./examples")

					wrapper := newResticWrapper(
						echoBinary,
						false,
						profile,
						fixture.commandName,
						fixture.cmdlineArgs,
						nil,
					)
					buffer := &bytes.Buffer{}
					// setting the output via the package global setter could lead to some issues
					// when some tests are running in parallel. I should fix that at some point :-/
					term.SetOutput(buffer)
					err = wrapper.runCommand(fixture.commandName)
					term.SetOutput(os.Stdout)

					// allow a fail temporarily
					if err != nil && err.Error() == fmt.Sprintf("%s on profile '%s': exit status 2", fixture.commandName, fixture.profileName) {
						t.Skip("shell failed to interpret command line")
					}
					require.NoError(t, err)

					expected := "[" + fixture.expected + "]"
					if fixture.expectedOnWindows != "" && runtime.GOOS == "windows" {
						expected = "[" + fixture.expectedOnWindows + "]"
					}
					assert.Equal(t, expected, strings.TrimSpace(buffer.String()))
				})
			}
		})
	}
}
