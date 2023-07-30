package main

//go:generate go install -v github.com/zyedidia/eget@latest
//go:generate eget vektra/mockery --upgrade-only --to "${GOPATH}/bin"
//go:generate mockery --config .mockery.yaml
