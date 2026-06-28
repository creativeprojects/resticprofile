package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise the *shell* completion scripts (contrib/completion/*) in a
// real bash, zsh and fish, rather than the Go Completer logic (see complete_test.go).
// They are integration tests: each one skips gracefully when the shell - or, for the
// restic delegation cases, restic itself - is not installed, so they never break a
// machine that lacks a given shell. They are also skipped in -short mode because the
// zsh harness drives a real pseudo-terminal and is therefore slow.

const testProfilesYAML = `
default:
  repository: /tmp/repo
  password-file: key
myprofile:
  repository: /tmp/repo2
  password-file: key2
mygroup:
  profiles:
    - default
    - myprofile
`

var (
	rpBuildOnce sync.Once
	rpBuildPath string
	errRPBuild  error
)

// resticprofileBinary builds the resticprofile binary once per test run and returns
// its path. The shell completion scripts shell out to this binary, so we need a real
// executable on disk.
func resticprofileBinary(t *testing.T) string {
	t.Helper()
	rpBuildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "rp-completion-bin")
		if err != nil {
			errRPBuild = err
			return
		}
		out := filepath.Join(dir, platform.Executable("resticprofile"))
		cmd := exec.CommandContext(t.Context(), "go", "build", "-o", out, ".")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			errRPBuild = err
			return
		}
		rpBuildPath = out
	})
	require.NoError(t, errRPBuild, "failed to build resticprofile binary")
	return rpBuildPath
}

// requireShell skips the test when the shell is not installed.
func requireShell(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s is not installed", name)
	}
	return path
}

// requireRestic skips the test when restic is not installed. restic's completion is
// offline (no repository, password or network needed), so the only requirement is the
// binary being on PATH.
func requireRestic(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("restic")
	if err != nil {
		t.Skip("restic is not installed (needed for restic delegation tests)")
	}
	return path
}

// completionEnv builds a temporary working directory holding the test config plus the
// resticprofile (and, when restic is present, restic) completion scripts for the given
// shell. It returns the directory and a PATH that puts the resticprofile binary first.
func completionEnv(t *testing.T, shell string, withRestic bool) (dir, path string) {
	t.Helper()
	rp := resticprofileBinary(t)
	dir = t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "profiles.yaml"), []byte(testProfilesYAML), 0o600))

	// resticprofile completion script for the shell
	writeGenerated(t, rp, dir, "_resticprofile_completion", "--"+shell+"-completion")

	if withRestic {
		restic := requireRestic(t)
		// restic ships its own completion via "restic generate"; "-" writes to stdout.
		out, err := exec.CommandContext(t.Context(), restic, "generate", "--"+shell+"-completion", "-").Output()
		require.NoError(t, err, "failed to generate restic %s completion", shell)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "_restic_completion"), out, 0o600))
	}

	// Put the freshly built resticprofile first on PATH, named "resticprofile".
	rpLink := filepath.Join(dir, platform.Executable("resticprofile"))
	require.NoError(t, copyFile(rp, rpLink))
	path = dir + string(os.PathListSeparator) + os.Getenv("PATH")
	return dir, path
}

func writeGenerated(t *testing.T, rp, dir, name, flag string) {
	t.Helper()
	out, err := exec.CommandContext(t.Context(), rp, "generate", flag).Output()
	require.NoError(t, err, "failed to generate completion for %s", flag)
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), out, 0o600))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

// --- fish -------------------------------------------------------------------

// fishComplete runs "complete --do-complete" for the given command line and returns the
// completion candidates (the part before the first tab, descriptions stripped).
func fishComplete(t *testing.T, dir, path, cmdline string, withRestic bool) []string {
	t.Helper()
	var src strings.Builder
	if withRestic {
		src.WriteString("source " + filepath.Join(dir, "_restic_completion") + "\n")
	}
	src.WriteString("source " + filepath.Join(dir, "_resticprofile_completion") + "\n")
	src.WriteString("complete --do-complete " + fishQuote(cmdline) + "\n")

	cmd := exec.CommandContext(t.Context(), "fish", "--no-config", "-c", src.String())
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "fish failed: %s", out)
	return firstColumn(string(out))
}

func TestFishShellCompletion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell completion test in short mode")
	}
	requireShell(t, "fish")
	dir, path := completionEnv(t, "fish", false)

	t.Run("own command", func(t *testing.T) {
		got := fishComplete(t, dir, path, "resticprofile prof", false)
		assert.Contains(t, got, "profiles")
	})
	t.Run("profile prefix", func(t *testing.T) {
		got := fishComplete(t, dir, path, "resticprofile def", false)
		assert.Contains(t, got, "default.")
	})
	t.Run("flag", func(t *testing.T) {
		got := fishComplete(t, dir, path, "resticprofile --comman", false)
		assert.Contains(t, got, "--command-output")
	})
	t.Run("restic delegation with profile prefix", func(t *testing.T) {
		dir, path := completionEnv(t, "fish", true)
		got := fishComplete(t, dir, path, "resticprofile default.", true)
		// every restic subcommand must be re-prefixed with the profile name
		assert.Contains(t, got, "default.backup")
		assert.Contains(t, got, "default.snapshots")
	})
	t.Run("restic command after a resticprofile flag", func(t *testing.T) {
		// "-n default" must not be forwarded to restic, otherwise it proposes nothing.
		dir, path := completionEnv(t, "fish", true)
		got := fishComplete(t, dir, path, "resticprofile -n default ba", true)
		assert.Contains(t, got, "backup")
	})
}

// --- bash -------------------------------------------------------------------

func requireBash4(t *testing.T) string {
	t.Helper()
	bash := requireShell(t, "bash")
	out, err := exec.CommandContext(t.Context(), bash, "-c", "echo ${BASH_VERSINFO[0]}").Output()
	require.NoError(t, err)

	major := strings.TrimSpace(string(out))
	if major == "0" || major == "1" || major == "2" || major == "3" {
		t.Skipf("bash 4+ required, found %s", major)
	}
	return bash
}

// bashComplete drives the _resticprofile completion function by setting COMP_WORDS /
// COMP_CWORD and reading COMPREPLY, the same way the interactive shell does.
func bashComplete(t *testing.T, dir, path string, words []string, withRestic bool) []string {
	t.Helper()
	var src strings.Builder
	src.WriteString("compopt() { :; }\n") // no-op outside interactive completion
	if withRestic {
		src.WriteString("source " + filepath.Join(dir, "_restic_completion") + "\n")
	}
	src.WriteString("source " + filepath.Join(dir, "_resticprofile_completion") + "\n")
	src.WriteString("COMP_WORDS=(")
	for _, w := range words {
		src.WriteString(bashQuote(w) + " ")
	}
	src.WriteString(")\n")
	src.WriteString("COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))\n")
	src.WriteString("COMP_LINE=\"${COMP_WORDS[*]}\"\n")
	src.WriteString("COMP_POINT=${#COMP_LINE}\n")
	src.WriteString("_resticprofile\n")
	src.WriteString("printf '%s\\n' \"${COMPREPLY[@]}\"\n")

	cmd := exec.CommandContext(t.Context(), "bash", "--norc", "--noprofile", "-c", src.String())
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "bash failed: %s", out)
	return nonEmptyLines(string(out))
}

func TestBashShellCompletion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell completion test in short mode")
	}
	requireBash4(t)
	dir, path := completionEnv(t, "bash", false)

	t.Run("own command", func(t *testing.T) {
		got := bashComplete(t, dir, path, []string{"resticprofile", "prof"}, false)
		assert.Contains(t, got, "profiles")
	})
	t.Run("profile prefix", func(t *testing.T) {
		got := bashComplete(t, dir, path, []string{"resticprofile", "def"}, false)
		assert.Contains(t, got, "default.")
	})
	t.Run("flag", func(t *testing.T) {
		got := bashComplete(t, dir, path, []string{"resticprofile", "--comman"}, false)
		assert.Contains(t, got, "--command-output")
	})
	t.Run("restic delegation with profile prefix", func(t *testing.T) {
		// restic's bash completion relies on the bash-completion framework
		// (_get_comp_words_by_ref etc.). Skip when it is not available.
		if !bashCompletionFrameworkAvailable() {
			t.Skip("bash-completion framework not installed")
		}
		dir, path := completionEnv(t, "bash", true)
		got := bashCompleteWithFramework(t, dir, path, []string{"resticprofile", "default."})
		assert.Contains(t, got, "default.backup")
		assert.Contains(t, got, "default.snapshots")
	})
	t.Run("restic command after a resticprofile flag", func(t *testing.T) {
		if !bashCompletionFrameworkAvailable() {
			t.Skip("bash-completion framework not installed")
		}
		// "-n default" must not be forwarded to restic, otherwise it proposes nothing.
		dir, path := completionEnv(t, "bash", true)
		got := bashCompleteWithFramework(t, dir, path, []string{"resticprofile", "-n", "default", "ba"})
		assert.Contains(t, got, "backup")
	})
}

func bashCompletionFrameworkPath() string {
	// The main framework file defines the helpers (_get_comp_words_by_ref etc.)
	// unconditionally. The profile.d wrappers bail out in a non-interactive shell,
	// so they are deliberately not used here.
	candidates := []string{
		"/usr/share/bash-completion/bash_completion",          // Linux (apt bash-completion)
		"/opt/homebrew/share/bash-completion/bash_completion", // macOS (brew bash-completion@2)
		"/usr/local/share/bash-completion/bash_completion",    // macOS Intel brew
		"/etc/bash_completion",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func bashCompletionFrameworkAvailable() bool { return bashCompletionFrameworkPath() != "" }

func bashCompleteWithFramework(t *testing.T, dir, path string, words []string) []string {
	t.Helper()
	var src strings.Builder
	src.WriteString("source " + bashCompletionFrameworkPath() + "\n")
	src.WriteString("compopt() { :; }\n")
	src.WriteString("source " + filepath.Join(dir, "_restic_completion") + "\n")
	src.WriteString("source " + filepath.Join(dir, "_resticprofile_completion") + "\n")
	src.WriteString("COMP_WORDS=(")
	for _, w := range words {
		src.WriteString(bashQuote(w) + " ")
	}
	src.WriteString(")\n")
	src.WriteString("COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))\n")
	src.WriteString("COMP_LINE=\"${COMP_WORDS[*]}\"\n")
	src.WriteString("COMP_POINT=${#COMP_LINE}\n")
	src.WriteString("_resticprofile\n")
	src.WriteString("printf '%s\\n' \"${COMPREPLY[@]}\"\n")

	cmd := exec.CommandContext(t.Context(), "bash", "--norc", "--noprofile", "-c", src.String())
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "bash failed: %s", out)
	return nonEmptyLines(string(out))
}

// --- zsh --------------------------------------------------------------------

// zshCompleteBuffer drives a real interactive zsh through a pseudo-terminal (zpty),
// types cmdline, presses TAB and then a bound widget that writes the resulting command
// line buffer to a file. Asserting on the completed buffer (rather than scraping the
// terminal listing) makes the result deterministic. The returned buffer is trimmed.
func zshCompleteBuffer(t *testing.T, dir, path, cmdline string, withRestic bool) string {
	t.Helper()
	return strings.TrimSpace(zshCompleteBufferRaw(t, dir, path, cmdline, withRestic))
}

// zshCompleteBufferRaw is like zshCompleteBuffer but preserves surrounding whitespace,
// so callers can assert on a trailing space (or its absence after a "profile." prefix).
func zshCompleteBufferRaw(t *testing.T, dir, path, cmdline string, withRestic bool) string {
	t.Helper()
	capFile := filepath.Join(dir, "buffer.txt")
	_ = os.Remove(capFile)

	harness := zshHarnessScript(dir, cmdline, capFile, withRestic)
	scriptPath := filepath.Join(dir, "harness.zsh")
	require.NoError(t, os.WriteFile(scriptPath, []byte(harness), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "zsh", "-f", scriptPath)
	// The inner zsh spawned by zpty inherits this PATH, so resticprofile resolves
	// without an in-pty "export PATH" (whose long echo can wrap and break the harness).
	cmd.Env = append(os.Environ(), "PATH="+path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zsh harness failed: %s", out)

	data, err := os.ReadFile(capFile)
	require.NoError(t, err, "zsh did not produce a completion buffer (output: %s)", out)
	// The buffer is captured as "BUF=[<buffer>]" so any trailing space is preserved.
	line := strings.TrimSpace(string(data)) // drop the trailing newline only
	line = strings.TrimPrefix(line, "BUF=")
	line = strings.TrimPrefix(line, "[")
	line = strings.TrimSuffix(line, "]")
	return line
}

// zshHarnessScript builds a script (run by an *outer* zsh) that drives an *inner*
// interactive zsh through a pseudo-terminal. Instead of relying on fixed sleeps, it
// synchronises on a readiness marker before typing the command line, and polls for the
// captured buffer file afterwards - both make the harness robust on slow CI machines.
// \$ escapes keep expansions for the inner zsh rather than the outer one.
func zshHarnessScript(dir, cmdline, capFile string, withRestic bool) string {
	resticSource := ""
	if withRestic {
		resticSource = `run "source ` + filepath.Join(dir, "_restic_completion") + `"` + "\n"
	}
	return `
zmodload zsh/zpty
drain() { local o; while zpty -r -t ZSH o 2>/dev/null; do :; done; }
# Read from the inner zsh until $1 is seen (~12s budget). Continuously reading also
# keeps the pty output buffer empty, otherwise the inner zsh blocks writing to it.
waitfor() {
    local acc="" chunk i
    for (( i = 0; i < 600; i++ )); do
        if zpty -r -t ZSH chunk 2>/dev/null; then
            acc+="$chunk"
            [[ "$acc" == *"$1"* ]] && return 0
        else
            sleep 0.02
        fi
    done
    return 1
}
# Run a setup command and wait until it has executed (a marker echoes back afterwards).
run() { zpty -w ZSH "$1"; zpty -w ZSH "print __RP_OK__"; waitfor "__RP_OK__"; }
zpty ZSH "exec zsh -f -i"
run ":"   # synchronise on the first prompt without matching its text
run "cd ` + dir + `; fpath=(` + dir + ` \$fpath)"
run "autoload -Uz compinit; compinit -u"
` + resticSource + `run "source ` + filepath.Join(dir, "_resticprofile_completion") + `"
run "_dump() { print -r -- \"BUF=[\$BUFFER]\" >! ` + capFile + ` }; zle -N _dump; bindkey '^Xb' _dump"
drain
# Type the command line, complete it with TAB, then dump the buffer with the widget.
# zsh reads input sequentially, so ^Xb is processed only after TAB completion returns.
zpty -w -n ZSH "` + cmdline + `"
zpty -w -n ZSH $'\t'
zpty -w -n ZSH $'\C-Xb'
# Poll for the captured buffer while draining (TAB may run restic, which spawns a
# subprocess, and may print a listing that must be drained to unblock the inner zsh).
for (( i = 0; i < 300; i++ )); do
    drain
    [[ -s "` + capFile + `" ]] && break
    sleep 0.05
done
zpty -d ZSH
`
}

func TestZshShellCompletion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping shell completion test in short mode")
	}
	requireShell(t, "zsh")
	dir, path := completionEnv(t, "zsh", false)

	t.Run("own command", func(t *testing.T) {
		got := zshCompleteBuffer(t, dir, path, "resticprofile prof", false)
		assert.Equal(t, "resticprofile profiles", got)
	})
	t.Run("profile prefix", func(t *testing.T) {
		got := zshCompleteBuffer(t, dir, path, "resticprofile myg", false)
		assert.Equal(t, "resticprofile mygroup.", got)
	})
	t.Run("profile prefix has no trailing space", func(t *testing.T) {
		// "<profile>." must not get a trailing space so it can be continued with a
		// command (e.g. "default.backup"), like the bash completion does.
		got := zshCompleteBufferRaw(t, dir, path, "resticprofile def", false)
		assert.Equal(t, "resticprofile default.", got)
	})
	t.Run("flag", func(t *testing.T) {
		got := zshCompleteBuffer(t, dir, path, "resticprofile --comman", false)
		assert.Equal(t, "resticprofile --command-output", got)
	})
	t.Run("flag of an own command", func(t *testing.T) {
		// A flag following an own command must complete (the words slice has to be
		// passed to resticprofile as separate arguments, not joined into one).
		got := zshCompleteBuffer(t, dir, path, "resticprofile status --a", false)
		assert.Equal(t, "resticprofile status --all", got)
	})
	t.Run("restic delegation re-prefixes the profile name", func(t *testing.T) {
		dir, path := completionEnv(t, "zsh", true)
		got := zshCompleteBuffer(t, dir, path, "resticprofile default.snap", true)
		assert.Equal(t, "resticprofile default.snapshots", got)
	})
	t.Run("restic delegation without prefix", func(t *testing.T) {
		dir, path := completionEnv(t, "zsh", true)
		got := zshCompleteBuffer(t, dir, path, "resticprofile bac", true)
		assert.Equal(t, "resticprofile backup", got)
	})
	t.Run("restic flag after a prefixed subcommand", func(t *testing.T) {
		// Exercises a multi-word slice through the restic delegation path.
		dir, path := completionEnv(t, "zsh", true)
		got := zshCompleteBuffer(t, dir, path, "resticprofile default.snapshots --ho", true)
		assert.Equal(t, "resticprofile default.snapshots --host", got)
	})
	t.Run("restic command after a resticprofile flag", func(t *testing.T) {
		// resticprofile's own flags (here "-n default") must not be forwarded to
		// restic, otherwise restic rejects them and proposes nothing.
		dir, path := completionEnv(t, "zsh", true)
		got := zshCompleteBuffer(t, dir, path, "resticprofile -n default ba", true)
		assert.Equal(t, "resticprofile -n default backup", got)
	})
}

// --- helpers ----------------------------------------------------------------

func firstColumn(out string) []string {
	var result []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		result = append(result, strings.SplitN(line, "\t", 2)[0])
	}
	return result
}

func nonEmptyLines(out string) []string {
	var result []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func fishQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
func bashQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
