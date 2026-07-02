package shell

import "strings"

type externalShellArgumentsComposer func(RunnerConfig, CommandConfig) []string

func getExternalShellComposer(shell string) externalShellArgumentsComposer {
	switch shell {
	case "sh", "bash", "zsh":
		return composeUnixShellArguments
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
