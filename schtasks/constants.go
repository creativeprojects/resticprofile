//go:build windows

package schtasks

const (
	dateFormat        = "2006-01-02T15:04:05-07:00"
	maxTriggers       = 60
	author            = "Author"
	taskSchema        = "http://schemas.microsoft.com/windows/2004/02/mit/task"
	taskSchemaVersion = "1.4"
	defaultPriority   = 8
	binaryPath        = "schtasks.exe"
	tasksPathPrefix   = `\resticprofile backup\`
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-string-format
	// O:owner_sid
	// G:group_sid
	// D:dacl_flags(string_ace1)(string_ace2)... (string_acen)  <---
	// S:sacl_flags(string_ace1)(string_ace2)... (string_acen)
	// With flag:
	// "AI"	SDDL_AUTO_INHERITED
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/ace-strings
	// - first field:
	// "A" 	SDDL_ACCESS_ALLOWED
	// - third field
	// "FA" 	SDDL_FILE_ALL 	FILE_GENERIC_ALL
	// "FR" 	SDDL_FILE_READ 	FILE_GENERIC_READ
	// "FW" 	SDDL_FILE_WRITE 	FILE_GENERIC_WRITE
	// "FX" 	SDDL_FILE_EXECUTE 	FILE_GENERIC_EXECUTE
	// From: https://learn.microsoft.com/en-us/windows/win32/secauthz/sid-strings
	// "AU" 	SDDL_AUTHENTICATED_USERS
	// "BA" 	SDDL_BUILTIN_ADMINISTRATORS
	// "LS" 	SDDL_LOCAL_SERVICE
	// "SY" 	SDDL_LOCAL_SYSTEM
	securityDescriptor = "D:AI(A;;FA;;;BA)(A;;FA;;;SY)(A;;FRFX;;;LS)(A;;FR;;;AU)"
)
