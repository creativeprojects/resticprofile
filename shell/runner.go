package shell

import (
	"context"
	"fmt"
	"strings"

	"github.com/creativeprojects/clog"
)

type Runner interface {
	Run(ctx context.Context, cmd CommandConfig) error
}

func NewRunner(config RunnerConfig, shells []string) (Runner, error) {
	var searchList []string
	for _, sh := range shells {
		if sh = strings.TrimSpace(sh); sh != "" {
			searchList = append(searchList, sh)
		}
	}
	if len(searchList) == 0 {
		searchList = getShellSearchList()
	}

	for _, sh := range searchList {
		config.Shell = Type(sh)

		runner, err := getRunner(config)
		if err != nil {
			clog.Warning(err)
			continue
		}
		return runner, nil
	}
	return nil, fmt.Errorf("cannot find a suitable shell after tyring %s", strings.Join(searchList, ", "))
}

func getRunner(config RunnerConfig) (Runner, error) {
	switch config.Shell {
	case TypeInternalPOSIX, TypeInternalBash:
		runner, err := NewInternalRunner(config)
		if err != nil {
			return nil, err
		}
		return runner, nil

	default:
		runner, err := NewExternalRunner(config)
		if err != nil {
			return nil, err
		}
		return runner, nil
	}
}
