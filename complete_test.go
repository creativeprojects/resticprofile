package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleter(t *testing.T) {
	completer := NewCompleter(ownCommands.All(), DefaultFlagsLoader)
	completer.init(nil)

	expectedProfiles := func() []string {
		return []string{"default", "full-backup", "linux", "no-cache", "root", "src", "stdin"}
	}

	newArgs := func(args ...string) (result []string) {
		result = append(result, "--config", "examples/profiles.conf")
		result = append(result, args...)
		return
	}

	// Flags and flag values
	t.Run("Flags", func(t *testing.T) {
		flagSet, _, _ := loadFlags(nil)

		hasFlags := false
		flagSet.VisitAll(func(*pflag.Flag) {
			hasFlags = true
		})
		require.True(t, hasFlags)

		// Flags
		t.Run("CompletesFlagSets", func(t *testing.T) {
			flagCompletion := func(flag *pflag.Flag, includeShorthand bool) (names []string) {
				if includeShorthand {
					names = append(names, fmt.Sprintf("-%s", flag.Shorthand))
				}
				names = append(names, fmt.Sprintf("--%s", flag.Name))
				return
			}

			t.Run("ReturnsAllFlagsExceptHidden", func(t *testing.T) {
				var expected []string
				flagSet.SortFlags = true
				flagSet.VisitAll(func(flag *pflag.Flag) {
					if !flag.Hidden {
						expected = append(expected, flagCompletion(flag, false)...)
					}
				})
				assert.Equal(t, expected, completer.completeFlagSet(""))
			})

			t.Run("ReturnsSpecificFlags", func(t *testing.T) {
				flagSet.VisitAll(func(flag *pflag.Flag) {
					expected := flagCompletion(flag, false)
					if flag.Hidden {
						expected = nil
					}
					actual := completer.completeFlagSet(fmt.Sprintf("--%s", flag.Name))
					assert.Subset(t, actual, expected)
					for _, flag := range actual {
						ok := slices.ContainsFunc(expected, func(prefix string) bool { return strings.HasPrefix(flag, prefix) })
						assert.True(t, ok, "prefixes not matched for %q", flag)
					}

					if len(flag.Shorthand) > 0 && !flag.Hidden {
						expected = flagCompletion(flag, true)[0:1]
						assert.Equal(t, expected, completer.completeFlagSet(fmt.Sprintf("-%s", flag.Shorthand)))
					}
				})
			})

			t.Run("RemovesAlreadySet", func(t *testing.T) {
				flagSet.VisitAll(func(flag *pflag.Flag) {
					completion := flagCompletion(flag, false)[0]
					completer.Complete(nil)
					if !flag.Hidden {
						assert.Contains(t, completer.completeFlagSet(""), completion)
					}
					completer.Complete(newArgs(fmt.Sprintf("--%s", flag.Name)))
					assert.NotContains(t, completer.completeFlagSet(""), completion)
				})
			})
		})

		// Flag values
		t.Run("CompletesFlagValues", func(t *testing.T) {

			t.Run("AllWithValues", func(t *testing.T) {
				flagSet.VisitAll(func(flag *pflag.Flag) {
					if flag.Hidden {
						return
					}

					completions := completer.completeFlagSetValue(flag, "")
					flagType := flag.Value.Type()
					if flagType == "duration" || flag.NoOptDefVal != "" {
						assert.Empty(t, completions, "Flag --%s", flag.Name)
					} else {
						assert.NotEmpty(t, completions, "Flag --%s", flag.Name)
					}
				})
			})

			testValues := func(flagName string, expected []string) func(t *testing.T) {
				return func(t *testing.T) {
					t.Run("ReturnsAllValues", func(t *testing.T) {
						actual := completer.Complete(newArgs(fmt.Sprintf("--%s", flagName), ""))
						assert.Equal(t, expected, actual)
					})

					t.Run("ReturnsSpecificValues", func(t *testing.T) {
						for _, value := range expected {
							actual := completer.completeFlagSetValue(flagSet.Lookup(flagName), value)
							assert.Equal(t, []string{value}, actual)
						}
					})
				}
			}

			t.Run("ConfigFlag", testValues("config", []string{RequestFileCompletion}))
			t.Run("FormatFlag", testValues("format", []string{"toml", "json", "yaml", "hcl"}))
			t.Run("LogFlag", testValues("log", []string{RequestFileCompletion}))
			t.Run("ThemeFlag", testValues("theme", []string{"dark", "light", "none"}))

			// Profiles from "examples/profiles.conf"
			t.Run("NameFlag", testValues("name", []string{
				"default", "full-backup", "linux", "no-cache", "root", "src", "stdin",
			}))
		})
	})

	// Profile completion
	t.Run("Profiles", func(t *testing.T) {
		const DevConfig = "examples/dev.conf"

		expectedProfiles := func() []string {
			return []string{"default", "full-backup", "repo-from-env", "rest", "root", "self", "src", "stdin"}
		}

		t.Run("Loading", func(t *testing.T) {
			t.Run("LoadsConfig", func(t *testing.T) {
				completer.Complete(newArgs("--config", DevConfig))
				assert.Equal(t, expectedProfiles(), completer.listProfileNames())
			})

			t.Run("RespectsFormat", func(t *testing.T) {
				completer.Complete(newArgs("--format", "json", "--config", DevConfig))
				assert.Equal(t, []string{}, completer.listProfileNames())
				completer.Complete(newArgs("--format", "toml", "--config", DevConfig))
				assert.Equal(t, expectedProfiles(), completer.listProfileNames())
			})

			t.Run("HandleNonExisting", func(t *testing.T) {
				completer.Complete(newArgs("--config", "non-existing-profiles.yaml"))
				assert.Equal(t, []string{}, completer.listProfileNames())
			})
		})

		t.Run("CompletesPrefix", func(t *testing.T) {
			t.Run("AllProfiles", func(t *testing.T) {
				completions := completer.Complete(newArgs("--config", DevConfig, ""))
				for _, profile := range expectedProfiles() {
					assert.Contains(t, completions, fmt.Sprintf("%s.", profile))
				}
			})

			t.Run("CommandsWithProfile", func(t *testing.T) {
				var commands []string
				for _, command := range ownCommands.All() {
					if command.needConfiguration && !command.hide && !command.hideInCompletion {
						commands = append(commands, command.name)
					}
				}

				require.NotEmpty(t, commands)
				sort.Strings(commands)

				for _, profile := range expectedProfiles() {
					prefix := fmt.Sprintf("%s.", profile)
					completions := completer.Complete(newArgs("--config", DevConfig, prefix))

					var expected []string
					for _, commandName := range commands {
						expected = append(expected, prefix+commandName)
					}
					expected = append(expected, prefix+RequestResticCompletion)

					assert.Equal(t, expected, completions)
				}
			})
		})
	})

	// Commands
	t.Run("Commands", func(t *testing.T) {
		var commands []string
		commandValues := map[string][]string{}

		for _, command := range ownCommands.All() {
			if !command.hide && !command.hideInCompletion {
				commands = append(commands, command.name)
			}
			for flag, _ := range command.flags {
				for _, v := range strings.Split(flag, ",") {
					v = strings.Split(strings.TrimSpace(v), " ")[0]
					commandValues[command.name] = append(commandValues[command.name], v)
					sort.Strings(commandValues[command.name])
				}
			}
		}

		require.NotEmpty(t, commands)
		sort.Strings(commands)

		t.Run("CompletesCommands", func(t *testing.T) {
			t.Run("AllVisibleCommands", func(t *testing.T) {
				assert.Equal(t, commands, completer.completeOwnCommands(""))
			})

			t.Run("SpecificCommands", func(t *testing.T) {
				for _, command := range commands {
					assert.Equal(t, []string{command}, completer.completeOwnCommands(command))
				}
			})
		})

		t.Run("CompletesCommandFlags", func(t *testing.T) {
			t.Run("AllFlags", func(t *testing.T) {
				for _, command := range commands {
					expected := commandValues[command]
					assert.Equal(t, expected, completer.Complete(newArgs(command, "-")), "Command %s", command)
				}
			})

			t.Run("SpecificFlags", func(t *testing.T) {
				for _, command := range commands {
					for _, flag := range commandValues[command] {
						assert.Equal(t, []string{flag}, completer.completeOwnCommandFlags(command, flag), "Command %s", command)
					}
				}
			})
		})
	})

	// Static test table on Complete()
	t.Run("Complete", func(t *testing.T) {
		tests := []struct {
			args     []string
			expected []string
		}{
			// Can complete by prefix
			{args: []string{"ful"}, expected: []string{"full-backup.", RequestResticCompletion}},
			{args: []string{"sched"}, expected: []string{"schedule", RequestResticCompletion}},
			{args: []string{"unsch"}, expected: []string{"unschedule", RequestResticCompletion}},
			{args: []string{"--nam"}, expected: []string{"--name"}},
			{args: []string{"--theme", "d"}, expected: []string{"dark"}},

			// Can complete shorthand
			{args: []string{"-c"}, expected: []string{RequestFileCompletion}},
			{args: []string{"-l"}, expected: []string{RequestFileCompletion}},
			{args: []string{"self-update", "-"}, expected: []string{"--quiet", "-q"}},
			{args: []string{"self-update", "-q"}, expected: nil},

			// Can completion commands after flags
			{args: []string{"--verbose", "schedule", "-"}, expected: []string{"--all", "--no-start"}},
			{args: []string{"--log", "file", "schedule", "-"}, expected: []string{"--all", "--no-start"}},

			// Flags are returned only once
			{args: []string{"--verb"}, expected: []string{"--verbose"}},
			{args: []string{"--verb", "--verb"}, expected: []string{"--verbose"}},
			{args: []string{"--verbose", "--verb"}, expected: nil},
			{args: []string{"schedule", "-"}, expected: []string{"--all", "--no-start"}},
			{args: []string{"schedule", "--all", "-"}, expected: []string{"--no-start"}},

			// Exact command match returns nothing (no duplication)
			{args: []string{"schedule"}, expected: nil},
			{args: []string{"unschedule"}, expected: nil},

			// Exact flag value match returns nothing (no duplication)
			{args: []string{"--name", "full-backup"}, expected: nil},
			{args: []string{"--theme", "light"}, expected: nil},
			{args: []string{"status", "--all"}, expected: nil},

			// Can set completion cursor position
			{args: []string{"__POS:1", "--log", "out.log", "--verbose", "schedule", "-"}, expected: []string{RequestFileCompletion}},
			{args: []string{"__POS:2", "--log", "out.log", "--verbose", "schedule", "-"}, expected: []string{RequestFileCompletion}},
			{args: []string{"__POS:4", "--log", "out.log", "--verbose", "schedule", "-"}, expected: nil},
			{args: []string{"__POS:4", "--log", "out.log", "--verbose", "schedule"}, expected: nil},
			{args: []string{"__POS:5", "--log", "out.log", "--verbose", "schedule", "-"}, expected: []string{"--all", "--no-start"}},
			{args: []string{"__POS:5", "--log", "out.log", "--verbose", "schedule"}, expected: []string{"--all", "--no-start"}},
			{args: []string{"__POS:INVALID", "--log", "out.log", "--verbose", "schedule", "-"}, expected: []string{"--all", "--no-start"}},
			{args: []string{"__POS:INVALID", "--log", "out.log", "--verbose", "schedule"}, expected: nil},

			// Unknown is delegated to restic
			{args: []string{"unknown-cmd"}, expected: []string{RequestResticCompletion}},
			{args: []string{"--verbose", "unknown-cmd"}, expected: []string{RequestResticCompletion}},
			{args: []string{"--log", "file", "unknown-cmd"}, expected: []string{RequestResticCompletion}},
			{args: []string{"--log", "file", "unknown-cmd", "--a-flag"}, expected: []string{RequestResticCompletion}},

			// Adds profile prefixes (for existing profiles, when name flag is not set)
			{args: []string{"unknown-profile.schedul"}, expected: []string{"schedule", RequestResticCompletion}},
			{args: []string{"full-backup.schedul"}, expected: []string{"full-backup.schedule", "full-backup." + RequestResticCompletion}},
			{args: []string{"--name", "something", "full-backup.schedul"}, expected: []string{"schedule", RequestResticCompletion}},
			{args: []string{"-n", "something", "full-backup.schedul"}, expected: []string{"schedule", RequestResticCompletion}},
			{args: []string{"full-backup.unknown-cmd"}, expected: []string{"full-backup." + RequestResticCompletion}},

			// Regression: Flag value completion does not complete flags matching value names
			{args: []string{"--log", "f"}, expected: []string{RequestFileCompletion}},
			{args: []string{"--log", "theme"}, expected: []string{RequestFileCompletion}},

			// Regression: Config loading respects flags when cursor position is specified
			{args: []string{"__POS:END", "--name"}, expected: expectedProfiles()},
			{args: []string{"__POS:END", "--format=json", "--name"}, expected: nil},
		}

		for index, test := range tests {
			name := fmt.Sprintf("#%d [%s]", index, strings.Join(test.args, ","))
			t.Run(name, func(t *testing.T) {
				completions := completer.Complete(newArgs(test.args...))
				assert.Equal(t, test.expected, completions)
			})
		}
	})
}
