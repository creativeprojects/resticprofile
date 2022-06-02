package shell

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConversionToArgs(t *testing.T) {
	args := NewArgs().GetAll()
	assert.Equal(t, []string{}, args)
}

func TestConversionToArgsFromFlags(t *testing.T) {
	args := NewArgs()
	args.AddFlags("aaa", []string{"one", "two"}, ArgConfigEscape)
	args.AddFlag("bbb", "three", ArgConfigEscape)
	assert.Equal(t, []string{"--aaa", "one", "--aaa", "two", "--bbb", "three"}, args.GetAll())
}

func TestConversionToArgsNoFlag(t *testing.T) {
	args := NewArgs()
	args.AddArgs([]string{"one", "two"}, ArgConfigEscape)
	args.AddArg("three", ArgConfigEscape)
	assert.Equal(t, []string{"one", "two", "three"}, args.GetAll())
}

func TestClone(t *testing.T) {
	args := NewArgs()
	args.AddFlag("x", "y", ArgConfigEscape)
	args.AddArg("more", ArgConfigEscape)
	args.SetLegacyArg(true)

	clone := args.Clone()
	assert.Equal(t, args.GetAll(), clone.GetAll())
	assert.Equal(t, args.ToMap(), clone.ToMap())
	assert.Equal(t, args.legacy, clone.legacy)

	assert.NotSame(t, args, clone)
	assert.NotSame(t, args.args, clone.args)
	assert.NotSame(t, args.args["x"], clone.args["x"])
	assert.NotSame(t, args.more, clone.more)
}

func TestWalk(t *testing.T) {
	args := NewArgs()
	args.AddFlag("x", "y", ArgConfigEscape)
	args.AddArg("more", ArgConfigEscape)

	var walked []string
	args.Walk(func(name string, arg *Arg) *Arg {
		walked = append(walked, arg.Value())
		if name == "x" {
			a := NewArg("newY", arg.Type())
			arg = &a
		}
		return arg
	})

	assert.Equal(t, []string{"y", "more"}, walked)
	assert.Equal(t, []string{"--x", "newY", "more"}, args.GetAll())
}

func TestConversionToArgs(t *testing.T) {
	args := NewArgs()
	args.AddFlags("aaa", []string{"simple", "with space", "with\"quote"}, ArgConfigEscape)
	args.AddFlags("bbb", []string{"simple", "with space", "with\"quote"}, ArgConfigKeepGlobQuote)
	args.AddArgs([]string{"with space", "with\"quote", "with$variable"}, ArgConfigEscape)
	args.AddArg("with space\"quote", ArgConfigKeepGlobQuote)
	args.AddArg("with$variable", ArgConfigKeepGlobQuote)

	expected := []string{
		"--aaa",
		"simple",
		"--aaa",
		`with\ space`,
		"--aaa",
		`with\"quote`,
		"--bbb",
		"simple",
		"--bbb",
		`"with space"`,
		"--bbb",
		`"with\"quote"`,
		`with\ space`,
		`with\"quote`,
		"with\\$variable",
		`"with space\"quote"`,
		"\"with$variable\"",
	}
	if runtime.GOOS == "windows" {
		expected = []string{
			"--aaa",
			"simple",
			"--aaa",
			"with space",
			"--aaa",
			"with\"quote",
			"--bbb",
			"simple",
			"--bbb",
			"with space",
			"--bbb",
			"with\"quote",
			"with space",
			"with\"quote",
			"with$variable",
			"with space\"quote",
			"with$variable",
		}
	}
	assert.Equal(t, expected, args.GetAll())
}

func TestPromoteSecondaryToPrimary(t *testing.T) {
	args := NewArgs()
	args.AddFlag("initialize", "true", ArgConfigEscape)
	args.AddFlag("repo", "first", ArgConfigEscape)          // replaced
	args.AddFlag("password-file", "key1", ArgConfigEscape)  // replaced
	args.AddFlag("key-hint", "key1", ArgConfigEscape)       // no replacement, but should be removed
	args.AddFlag("repo2", "second", ArgConfigEscape)        // promoted to repo
	args.AddFlag("password-file2", "key2", ArgConfigEscape) // promoted to password-file
	args.AddFlag("other2", "keep", ArgConfigEscape)         // should stay

	args.PromoteSecondaryToPrimary(false)
	result := args.ToMap()
	assert.Equal(t, map[string][]string{
		"initialize":    {"true"},
		"password-file": {"key2"},
		"repo":          {"second"},
		"other2":        {"keep"},
	}, result)
}

func TestSwapSecondaryWithPrimary(t *testing.T) {
	args := NewArgs()
	args.AddFlag("initialize", "true", ArgConfigEscape)
	args.AddFlag("repo", "first", ArgConfigEscape)                // promoted to repo2
	args.AddFlag("password-file", "key1", ArgConfigEscape)        // promoted to password-file2
	args.AddFlag("key-hint", "key1", ArgConfigEscape)             // promoted to key-hint2
	args.AddFlag("repo2", "second", ArgConfigEscape)              // promoted to repo
	args.AddFlag("password-file2", "key2", ArgConfigEscape)       // promoted to password-file
	args.AddFlag("password-command2", "command", ArgConfigEscape) // promoted to password-command
	args.AddFlag("other2", "keep", ArgConfigEscape)               // should stay the same

	args.PromoteSecondaryToPrimary(true)
	result := args.ToMap()
	assert.Equal(t, map[string][]string{
		"initialize":       {"true"},
		"password-file":    {"key2"},
		"repo":             {"second"},
		"password-command": {"command"},
		"password-file2":   {"key1"},
		"repo2":            {"first"},
		"key-hint2":        {"key1"},
		"other2":           {"keep"},
	}, result)
}
