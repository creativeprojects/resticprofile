package schtasks

// The identifier that is used to specify the privilege level that is required to run the tasks that are associated with the principal.
// https://learn.microsoft.com/en-us/windows/win32/taskschd/taskschedulerschema-runleveltype-simpletype
type RunLevel string

const (
	RunLevelLeastPrivilege RunLevel = "LeastPrivilege"
	RunLevelHighest        RunLevel = "HighestAvailable"
)
