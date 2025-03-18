//go:build windows

package user

import (
	"os/user"
)

type User struct {
	Uid      int
	Gid      int
	Username string
	SudoRoot bool
}

func Current() User {
	username := ""
	current, err := user.Current()
	if err == nil {
		username = current.Username
	}
	return User{
		Uid:      -1,
		Gid:      -1,
		Username: username,
		SudoRoot: false,
	}
}
