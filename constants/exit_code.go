package constants

const (
	ExitSuccess = iota
	ExitGeneralError
	ExitErrorInvalidFlags
	ExitRunningOnBattery
	ExitCannotSetupRemoteConfiguration
	ExitErrorChildHasNoParentPort = 10
)
