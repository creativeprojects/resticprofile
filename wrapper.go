package main

import (
	"strings"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
)

type resticWrapper struct {
	resticBinary string
	profile      *config.Profile
	moreArgs     []string
}

func newResticWrapper(resticBinary string, profile *config.Profile, moreArgs []string) *resticWrapper {
	return &resticWrapper{
		resticBinary: resticBinary,
		profile:      profile,
		moreArgs:     moreArgs,
	}
}

func (r *resticWrapper) initialize() {

}

func (r *resticWrapper) cleanup() {

}

func (r *resticWrapper) check() {

}

func (r *resticWrapper) command(command string) error {
	arguments := append([]string{command}, r.moreArgs...)
	clog.Debugf("Starting command: %s %s", r.resticBinary, strings.Join(arguments, " "))
	rCommand := newCommand(r.resticBinary, arguments, nil)
	return runCommand(rCommand)
}
