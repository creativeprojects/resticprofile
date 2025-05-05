//go:build windows

package user

import (
	"os/user"
)

func Current() User {
	var username, homedir string
	current, err := user.Current()
	if err == nil {
		username = current.Username
		homedir = current.HomeDir
	}
	return User{
		Uid:         -1,
		Gid:         -1,
		Username:    username,
		UserHomeDir: homedir,
		Sudo:        false,
	}
}
