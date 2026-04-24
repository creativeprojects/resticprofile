package constants

const (
	ExitSuccess = iota
	ExitGeneralError
	ExitErrorInvalidFlags
	ExitRunningOnBattery
	ExitCannotSetupRemoteConfiguration
	ExitNotEnoughMemory
	ExitResticBinaryNotFound
	ExitErrorChildHasNoParentPort = 10
)
