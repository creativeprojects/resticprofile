package schedule

import (
	"testing"
)

func TestRawArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"empty args", []string{}},
		{"simple args", []string{"arg1", "arg2"}},
		{"args with spaces", []string{"C:\\Program Files\\app.exe", "--config", "C:\\My Documents\\config.toml"}},
		{"args with special chars", []string{"--name", "my-task!", "--config=test.conf"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := NewCommandArguments(tt.args)
			rawArgs := ca.RawArgs()
			if len(rawArgs) != len(tt.args) {
				t.Errorf("expected %d raw arguments, got %d", len(tt.args), len(rawArgs))
			}
			for i, arg := range tt.args {
				if rawArgs[i] != arg {
					t.Errorf("expected raw argument %d to be %s, got %s", i, arg, rawArgs[i])
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"arg1"}, "arg1"},
		{[]string{"arg1 with space"}, `"arg1 with space"`},
		{[]string{"arg1", "arg2"}, "arg1 arg2"},
		{[]string{"arg1", "arg with spaces"}, `arg1 "arg with spaces"`},
		{[]string{"arg1", "arg with spaces", "anotherArg"}, `arg1 "arg with spaces" anotherArg`},
		{[]string{"--config", "C:\\Program Files\\config.toml"}, `--config "C:\Program Files\config.toml"`},
		{[]string{"--config", "C:\\Users\\John Doe\\Documents\\config.toml", "--name", "backup task"},
			`--config "C:\Users\John Doe\Documents\config.toml" --name "backup task"`},
		{[]string{"--config", "C:\\My Files\\config.toml", "--no-ansi"},
			`--config "C:\My Files\config.toml" --no-ansi`},
	}

	for _, test := range tests {
		ca := NewCommandArguments(test.args)
		result := ca.String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func TestConfigFile(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"--config"}, ""},
		{[]string{"--config", "config.toml"}, "config.toml"},
		{[]string{"--config", "C:\\Program Files\\config.toml"}, "C:\\Program Files\\config.toml"},
		{[]string{"--name", "backup", "--config", "config.toml"}, "config.toml"},
		{[]string{"--config", "config.toml", "--name", "backup"}, "config.toml"},
		{[]string{"--name", "backup", "--config", "config.toml", "--no-ansi"}, "config.toml"},
		{[]string{"--name", "backup", "--no-ansi", "--config", "config.toml"}, "config.toml"},
		{[]string{"--name", "backup", "--no-ansi"}, ""},
	}

	for _, test := range tests {
		ca := NewCommandArguments(test.args)
		result := ca.ConfigFile()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}
