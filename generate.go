package main

//go:generate go install -v github.com/zyedidia/eget@latest
//go:generate "${GOPATH}/bin/eget" vektra/mockery --upgrade-only --to "${GOPATH}/bin"
//go:generate "${GOPATH}/bin/mockery" --config .mockery.yaml
