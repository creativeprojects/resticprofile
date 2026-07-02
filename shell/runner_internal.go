package shell

import (
	"bytes"
	"context"
	"io"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type InternalRunner struct {
	parser *syntax.Parser
	runner *interp.Runner
}

func NewInternalRunner(config RunnerConfig) (*InternalRunner, error) {
	runnerType := syntax.LangPOSIX
	if config.Shell == TypeInternalBash {
		runnerType = syntax.LangBash
	}
	parser := syntax.NewParser(syntax.Variant(runnerType))

	env := NewEnv().AddEnviron(config.Env)
	runner, err := interp.New(
		interp.Env(env),
		interp.StdIO(config.Stdin, config.Stdout, config.Stderr),
		interp.Dir(config.Dir),
	)
	if err != nil {
		return nil, err
	}
	return &InternalRunner{
		parser: parser,
		runner: runner,
	}, nil
}

func (r *InternalRunner) Run(ctx context.Context, cmd CommandConfig) error {
	script, err := r.parser.Parse(buildCommandBuffer(cmd), "")
	if err != nil {
		return err
	}
	err = r.runner.Run(ctx, script)
	if err != nil {
		return err
	}
	return nil
}

func buildCommandBuffer(cmd CommandConfig) io.Reader {
	buffer := new(bytes.Buffer)
	buffer.WriteString(cmd.Command)
	for _, arg := range cmd.Args {
		_ = buffer.WriteByte(' ')
		_, _ = buffer.WriteString(arg)
	}
	return buffer
}
