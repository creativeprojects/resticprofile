package schtasks

const serviceAccount = "S-1-5-18"

type Principals struct {
	Principal Principal `xml:"Principal"`
}

type Principal struct {
	ID        string    `xml:"id,attr"`
	UserId    string    `xml:"UserId"`
	LogonType LogonType `xml:"LogonType,omitempty"`
	RunLevel  RunLevel  `xml:"RunLevel,omitempty"`
}

// LongType specifies the security logon method required to run those tasks associated with the principal.
// https://learn.microsoft.com/en-us/windows/win32/taskschd/taskschedulerschema-logontype-principaltype-element
type LogonType string

const (
	LogonTypeServiceForUser   LogonType = "S4U"              // User must log on using a service for user (S4U) logon. When an S4U logon is used, no password is stored by the system and there is no access to the network or encrypted files.
	LogonTypePassword         LogonType = "Password"         // User must log on using a password.
	LogonTypeInteractiveToken LogonType = "InteractiveToken" // User must already be logged on. The task will be run only in an existing interactive session.
)

// The identifier that is used to specify the privilege level that is required to run the tasks that are associated with the principal.
// https://learn.microsoft.com/en-us/windows/win32/taskschd/taskschedulerschema-runleveltype-simpletype
type RunLevel string

const (
	RunLevelDefault        RunLevel = ""
	RunLevelLeastPrivilege RunLevel = "LeastPrivilege"
	RunLevelHighest        RunLevel = "HighestAvailable"
)
