package shell

import "github.com/creativeprojects/resticprofile/mask"

type ArgOption interface {
	setup(arg *Arg)
}

// EmptyArgOption is an option to create an specifically empty argument (e.g. --flag="")
type EmptyArgOption struct{}

func (o *EmptyArgOption) setup(arg *Arg) {
	arg.empty = true
}

type ConfidentialArgOption struct {
	ConfidentialFilter filterFunc
}

func NewConfidentialArgOption(confidential bool) *ConfidentialArgOption {
	var confidentialFilter filterFunc
	if confidential {
		confidentialFilter = func(value string) string {
			return mask.Submatches(mask.RepositoryConfidentialPart, value)
		}
	}
	return &ConfidentialArgOption{
		ConfidentialFilter: confidentialFilter,
	}
}

func (o *ConfidentialArgOption) setup(arg *Arg) {
	arg.confidentialFilter = o.ConfidentialFilter
}
