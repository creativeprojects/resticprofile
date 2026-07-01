package shell

type InternalShellTransformer struct{}

func (t InternalShellTransformer) Transform(name string, arg Arg) ([]Arg, bool) {
	return []Arg{arg}, false
}

var _ ArgTransformer = &InternalShellTransformer{}
