package user

type User struct {
	Uid      int
	Gid      int
	Username string
	HomeDir  string
	SudoRoot bool
}

// IsRoot returns true when the user has used sudo to run the command,
// or if the user was simply logged on a root
func (u User) IsRoot() bool {
	return u.SudoRoot || u.Uid == 0
}
