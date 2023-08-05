package shell

import (
	"github.com/creativeprojects/resticprofile/constants"
)

func GetOutputScanner(commandName string, jsonOutput bool) ScanOutput {
	switch commandName {
	case constants.CommandBackup:
		if jsonOutput {
			return scanBackupJson
		}
		return scanBackupPlain
	}
	return nil
}
