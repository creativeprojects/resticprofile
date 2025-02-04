package shell

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConversionToArgs(t *testing.T) {
	t.Parallel()

	args := NewArgs().GetAll()
	assert.Equal(t, []string{}, args)
}

func TestConversionToArgsFromFlags(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlags("aaa", NewArgsSlice([]string{"one", "two"}, ArgConfigEscape))
	args.AddFlag("bbb", NewArg("three", ArgConfigEscape))
	assert.Equal(t, []string{"--aaa=one", "--aaa=two", "--bbb=three"}, args.GetAll())
}

func TestConversionToArgsNoFlag(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddArgs(NewArgsSlice([]string{"one", "two"}, ArgConfigEscape))
	args.AddArg(NewArg("three", ArgConfigEscape))
	assert.Equal(t, []string{"one", "two", "three"}, args.GetAll())
}

func TestClone(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlag("x", NewArg("y", ArgConfigEscape))
	args.AddArg(NewArg("more", ArgConfigEscape))

	clone := args.Clone()
	assert.Equal(t, args.GetAll(), clone.GetAll())
	assert.Equal(t, args.ToMap(), clone.ToMap())

	assert.NotSame(t, args, clone)
	assert.NotSame(t, &args.args, &clone.args)
	assert.NotSame(t, &args.more, &clone.more)
}

func TestWalk(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlag("x", NewArg("y", ArgConfigEscape))
	args.AddArg(NewArg("more", ArgConfigEscape))

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
	assert.Equal(t, []string{"--x=newY", "more"}, args.GetAll())
}

func TestRenameAndRemove(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlag("x", NewArg("y", ArgConfigEscape))
	args.AddArg(NewArg("more", ArgConfigEscape))

	args.Rename("more", "new-more")
	args.Rename("x", "new-x")
	assert.ElementsMatch(t, []string{"new-more", "--new-x=y"}, args.GetAll())

	args.RemoveArg("new-more")
	assert.Equal(t, []string{"--new-x=y"}, args.GetAll())

	args.Remove("new-x")
	assert.Empty(t, args.GetAll())
}

func TestConversionToArgs(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlags("aaa", NewArgsSlice([]string{"simple", "with space", "with\"quote"}, ArgConfigEscape))
	args.AddFlags("bbb", NewArgsSlice([]string{"simple", "with space", "with\"quote"}, ArgConfigKeepGlobQuote))
	args.AddArgs(NewArgsSlice([]string{"with space", "with\"quote", "with$variable"}, ArgConfigEscape))
	args.AddArg(NewArg("with space\"quote", ArgConfigKeepGlobQuote))
	args.AddArg(NewArg("with$variable", ArgConfigKeepGlobQuote))

	expected := []string{
		"--aaa=simple",
		`--aaa=with\ space`,
		`--aaa=with\"quote`,
		"--bbb=simple",
		`--bbb="with space"`,
		`--bbb="with\"quote"`,
		`with\ space`,
		`with\"quote`,
		"with\\$variable",
		`"with space\"quote"`,
		"\"with$variable\"",
	}
	if runtime.GOOS == "windows" {
		expected = []string{
			"--aaa=simple",
			"--aaa=with space",
			"--aaa=with\"quote",
			"--bbb=simple",
			"--bbb=with space",
			"--bbb=with\"quote",
			"with space",
			"with\"quote",
			"with$variable",
			"with space\"quote",
			"with$variable",
		}
	}
	assert.Equal(t, expected, args.GetAll())
}

type testModifier struct{}

func (m testModifier) Arg(name string, arg *Arg) (*Arg, bool) {
	newArg := arg.Clone()
	newArg.value = "modified-" + arg.Value()
	return &newArg, true
}

func TestArgsModify(t *testing.T) {
	t.Parallel()

	args := &Args{
		args: map[string][]Arg{
			"aaa": {NewArg("value", ArgConfigEscape)},
			"bbb": {NewArg("value", ArgConfigKeepGlobQuote)},
		},
		more: []Arg{
			NewArg("value", ArgConfigEscape),
			NewArg("value", ArgConfigKeepGlobQuote),
		},
	}

	newArgs := args.Modify(testModifier{})

	expected := []string{"--aaa=modified-value", "--bbb=modified-value", "modified-value", "modified-value"}
	assert.Equal(t, expected, newArgs.GetAll())
}
