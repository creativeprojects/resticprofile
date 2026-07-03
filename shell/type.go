package shell

type Type string

const (
	TypeNoShell           Type = "no-shell"
	TypeInternalPOSIX     Type = "internal:sh"
	TypeInternalBash      Type = "internal:bash"
	TypeInternalZsh       Type = "internal:zsh"
	TypeExternalPOSIX     Type = "sh"
	TypeExternalBash      Type = "bash"
	TypeExternalZsh       Type = "zsh"
	TypeWindowsCmd        Type = "cmd.exe"
	TypeWindowsPowershell Type = "powershell.exe"
)
