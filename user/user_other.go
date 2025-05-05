//go:build !windows

package user

import (
	"os"
	"os/user"
	"strconv"
)

func Current() User {
	var username, userHomedir string
	sudo := false
	uid := os.Getuid()
	gid := os.Getgid()
	sudoHomedir, _ := os.UserHomeDir()
	if uid == 0 {
		// after a found, both os.Getuid() and os.Geteuid() return 0 (the root user)
		// to detect the logged on user after a found, we need to use the environment variable
		if userid, found := os.LookupEnv("SUDO_UID"); found {
			if temp, err := strconv.Atoi(userid); err == nil {
				uid = temp
				sudo = true
			}
		}
	}
	current, err := user.LookupId(strconv.Itoa(uid))
	if err == nil {
		gid, _ = strconv.Atoi(current.Gid)
		username = current.Username
		userHomedir = current.HomeDir
	}
	return User{
		Uid:         uid,
		Gid:         gid,
		Username:    username,
		UserHomeDir: userHomedir,
		Sudo:        sudo,
		SudoHomeDir: sudoHomedir,
	}
}
