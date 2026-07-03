package shell

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/creativeprojects/resticprofile/util/collect"
)

type externalShellArgumentsComposer func(RunnerConfig, CommandConfig) []string

func getExternalShellComposer(shell string) externalShellArgumentsComposer {
	switch shell {
	case "sh", "bash", "zsh":
		return composeUnixShellArguments
	case "cmd", "cmd.exe":
		return composeWindowsCmdArguments2
	case "powershell", "powershell.exe":
		return composePowershellArguments1
	default:
		return nil
	}
}

func composeUnixShellArguments(_ RunnerConfig, cmdConfig CommandConfig) []string {
	// Flatten all arguments into one string, sh and bash expects one big string
	command := resolveCommand(cmdConfig.Command)
	flatCommand := strings.Join(append([]string{command}, cmdConfig.Args...), " ")

	return []string{
		"-c",
		flatCommand,
	}
}

func composeWindowsCmdArguments2(_ RunnerConfig, cmdConfig CommandConfig) []string {
	// Rewrite unix style env variables ($var) to delayed cmd style (!var!)
	mapper := func(name string) string { return fmt.Sprintf("!%s!", name) }
	arguments := rewriteVariables(cmdConfig.Args, mapper)
	command := resolveCommand(rewriteVariables([]string{cmdConfig.Command}, mapper)[0])

	// Enable delayed variable expansion "/V:ON" to support !variable! syntax
	return append(
		[]string{"/V:ON", "/C", command},
		removeQuotes(arguments)...,
	)
}

func composePowershellArguments1(runnerConfig RunnerConfig, cmdConfig CommandConfig) []string {
	commandEnvNames := collect.From(runnerConfig.Env, func(env string) string {
		return strings.ToLower(strings.SplitN(env, "=", 2)[0])
	})

	// Rewrite unix style env variables ($var) to powershell env style ($Env:var) with fallback to local variable
	// Is limited to existing env variables and will not modify variables that have been set to a value (using $var=).
	mapper := func(name string) string {
		hasEnv := slices.Contains(commandEnvNames, strings.ToLower(name))
		if !hasEnv {
			_, hasEnv = os.LookupEnv(name)
		}

		if hasEnv && !powershellBuiltins.MatchString(name) {
			assignment, err := regexp.Compile(fmt.Sprintf(`(?i)\$%[1]s\s*=|-Variable\s+-Name\s+"%[1]s"`, regexp.QuoteMeta(name)))
			if err != nil || !assignment.MatchString(cmdConfig.Command) {
				return fmt.Sprintf("${Env:%s}", name)
			}
		}

		return "" // leave variable unchanged
	}

	// Check if mapper is disabled (when $RESTICPROFILE_PWSH_NO_AUTOENV is set to any value)
	if mapper("RESTICPROFILE_PWSH_NO_AUTOENV") != "" {
		mapper = func(name string) string { return "" }
	}

	arguments := rewriteVariables(cmdConfig.Args, mapper)
	command := resolveCommand(rewriteVariables([]string{cmdConfig.Command}, mapper)[0])

	// Absolute path to an executable must be escaped in PWSH:
	if strings.Contains(command, " ") && !strings.Contains(command, `"`) {
		if stat, err := os.Stat(command); err == nil && !stat.IsDir() {
			command = fmt.Sprintf(`& "%s"`, command)
		}
	}

	// Arguments are treated as markup in PWSH, we need to escape them as strings
	arguments = collect.From(removeQuotes(arguments), func(arg string) string {
		if !strings.Contains(arg, "`") {
			arg = fmt.Sprintf(`"%s"`, strings.ReplaceAll(arg, "\"", "`\""))
		}
		return arg
	})

	return append(
		[]string{"-Command", command},
		arguments...,
	)
}
