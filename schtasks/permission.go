package schtasks

import (
	"fmt"
	"os/user"

	"github.com/creativeprojects/resticprofile/term"
)

// Permission is a choice between System, User and User Logged On
type Permission int

// Permission available
const (
	UserAccount Permission = iota
	SystemAccount
	UserLoggedOnAccount
)

var (
	// current user
	userName = ""
	// ask the user password only once
	userPassword = ""
)

// userCredentials asks for the user password only once, and keeps it in cache
func userCredentials() (string, string, error) {
	if userName != "" {
		// we've been here already: we don't check for blank password as it's a valid password
		return userName, userPassword, nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", "", err
	}
	userName = currentUser.Username

	fmt.Printf("\nCreating task for user %s\n", userName)
	fmt.Printf("Task Scheduler requires your Windows password to validate the task: ")
	userPassword, err = term.ReadPassword()
	if err != nil {
		return "", "", err
	}
	return userName, userPassword, nil
}
