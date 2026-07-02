package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/creativeprojects/clog"
	"mvdan.cc/sh/v3/expand"
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
	clog.Debugf("using internal shell with %s syntax", runnerType.String())
	parser := syntax.NewParser(syntax.Variant(runnerType))

	env := NewEnv().AddEnviron(config.Env)
	runner, err := interp.New(
		interp.Env(env),
		interp.StdIO(config.Stdin, config.Stdout, config.Stderr),
		interp.Dir(config.Dir),
		interp.ExecHandlers(execHandler),
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

func execHandler(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		hc := interp.HandlerCtx(ctx)
		path, err := interp.LookPathDir(hc.Dir, hc.Env, args[0])
		if err != nil {
			fmt.Fprintln(hc.Stderr, err)
			return interp.ExitStatus(127)
		}
		cmd := exec.Cmd{
			Path:   path,
			Args:   args,
			Env:    execEnv(hc.Env),
			Dir:    hc.Dir,
			Stdin:  hc.Stdin,
			Stdout: hc.Stdout,
			Stderr: hc.Stderr,
		}

		err = cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}

		if target, ok := errors.AsType[*exec.ExitError](err); ok {
			if status, ok := target.Sys().(waitStatus); ok && status.Signaled() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return interp.ExitStatus(128 + status.Signal()) //nolint:gosec
			}
			return interp.ExitStatus(target.ExitCode()) //nolint:gosec
		}

		if target, ok := errors.AsType[*exec.Error](err); ok {
			// did not start
			fmt.Fprintf(hc.Stderr, "%v\n", target)
			return interp.ExitStatus(127)
		}

		return err
	}
}

func execEnv(env expand.Environ) []string {
	list := make([]string, 0, 64)
	for name, vr := range env.Each {
		if !vr.IsSet() {
			// If a variable is set globally but unset in the runner,
			// we need to ensure it's not part of the final list.
			for i, kv := range list {
				if strings.HasPrefix(kv, name+"=") {
					list[i] = ""
				}
			}
		}
		if vr.Exported && vr.Kind == expand.String {
			list = append(list, name+"="+vr.String())
		}
	}
	return list
}
