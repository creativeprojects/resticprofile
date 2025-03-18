//go:build !darwin

package user

import (
	"os"
	"os/user"
	"strconv"
)

type User struct {
	Uid      int
	Gid      int
	Username string
	SudoRoot bool
}

func Current() User {
	username := ""
	uid := os.Getuid()
	gid := os.Getgid()

	current, err := user.Current()
	if err == nil {
		uid, _ = strconv.Atoi(current.Uid)
		gid, _ = strconv.Atoi(current.Gid)
		username = current.Username
	}
	return User{
		Uid:      uid,
		Gid:      gid,
		Username: username,
		SudoRoot: false,
	}
}
