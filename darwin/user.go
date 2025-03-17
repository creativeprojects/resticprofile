//go:build darwin

package darwin

import (
	"os"
	"os/user"
	"strconv"
)

type User struct {
	Uid      int
	Gid      int
	SudoRoot bool
}

func CurrentUser() User {
	sudoed := false
	uid := os.Getuid()
	gid := os.Getgid()
	if uid == 0 {
		// after a sudo, macOs returns the root user on both os.Getuid() and os.Geteuid()
		// to detect the logged on user after a sudo, we need to use the environment variable
		if userid, sudo := os.LookupEnv("SUDO_UID"); sudo {
			if temp, err := strconv.Atoi(userid); err == nil {
				uid = temp
				sudoed = true
			}
		}
		current, err := user.LookupId(strconv.Itoa(uid))
		if err == nil {
			gid, _ = strconv.Atoi(current.Gid)
		}
	}
	return User{
		Uid:      uid,
		Gid:      gid,
		SudoRoot: sudoed,
	}
}
