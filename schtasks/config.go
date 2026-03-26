//go:build windows

package schtasks

type Config struct {
	ProfileName        string
	CommandName        string
	Command            string
	Arguments          string
	WorkingDirectory   string
	JobDescription     string
	RunLevel           string
	StartWhenAvailable bool
}
