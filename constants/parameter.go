package constants

// Parameter
const (
	ParameterIONice            = "ionice"
	ParameterIONiceClass       = "ionice-class"
	ParameterIONiceLevel       = "ionice-level"
	ParameterNice              = "nice"
	ParameterPriority          = "priority"
	ParameterDefaultCommand    = "default-command"
	ParameterInitialize        = "initialize"
	ParameterResticBinary      = "restic-binary"
	ParameterInherit           = "inherit"
	ParameterHost              = "host"
	ParameterPath              = "path"
	ParameterTag               = "tag"
	ParameterDescription       = "description"
	ParameterVersion           = "version"
	ParameterCopyChunkerParams = "copy-chunker-params"
)

// Parameter that has a "2" after its name when needing 2 repositories (init and copy only)
var (
	SwappableParameters = []string{
		"repo",
		"repository-file",
		"password-file",
		"password-command",
		"key-hint",
	}
)
