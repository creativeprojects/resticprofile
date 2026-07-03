package shell

type RunnerConfig struct {
	DryRun bool
	Shell  Type
	Env    []string
	Dir    string
}
