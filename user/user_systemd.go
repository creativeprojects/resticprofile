//go:build !windows && !darwin

package user

import (
	"errors"
	"fmt"
	"os"
)

func (u User) HasLingering() bool {
	_, err := os.Stat(fmt.Sprintf("/var/lib/systemd/linger/%s", u.Username))
	return err == nil || !errors.Is(err, os.ErrNotExist)
}
