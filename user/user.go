package user

type User struct {
	Uid         int
	Gid         int
	Username    string
	UserHomeDir string
	Sudo        bool
	SudoHomeDir string
}

// IsRoot returns true when the user has used sudo to run the command,
// or if the user was simply logged on a root
func (u User) IsRoot() bool {
	return u.Sudo || u.Uid == 0
}
