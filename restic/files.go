package restic

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
)

const (
	commandsJsonFile = "commands.json"
	resticPGPKey     = "pgp-key.asc"
)

//go:embed *.json *.asc
var eFS embed.FS

func openEmbedded(filename string) (fs.File, error) {
	return eFS.Open(filename)
}

func readResticPGPKey() (keyBytes []byte) {
	key, err := openEmbedded(resticPGPKey)
	if err == nil {
		defer func() { _ = key.Close() }()
		keyBytes, err = io.ReadAll(key)
	}
	if err != nil {
		panic(err)
	}
	return
}

func LoadEmbeddedCommands() {
	file, err := openEmbedded(commandsJsonFile)
	if err == nil {
		defer file.Close()

		var cmds map[string]*command
		if cmds, err = loadCommandsFromReader(file); err == nil {
			ClearCommands()
			commands = cmds
		}

		for _, goos := range []string{"windows"} {
			if err == nil {
				err = loadEmbeddedOSExtensions(goos, cmds)
			}
		}
	}
	if err != nil {
		panic(err)
	}
}

func loadEmbeddedOSExtensions(goos string, cmds map[string]*command) error {
	ext := path.Ext(commandsJsonFile)
	base := strings.TrimSuffix(commandsJsonFile, ext)
	extensionsFile := fmt.Sprintf("%s_%s%s", base, goos, ext)

	file, err := openEmbedded(extensionsFile)
	if err == nil {
		defer file.Close()

		var extensions map[string][]Option
		if extensions, err = loadCommandExtensionsFromReader(file); err == nil {
			for _, options := range extensions {
				for i, _ := range options {
					if !slices.Contains(options[i].OnlyInOS, goos) {
						options[i].OnlyInOS = append(options[i].OnlyInOS, goos)
					}
				}
			}
			applyCommandExtensions(cmds, extensions)
		}
	}
	return err
}

func init() {
	LoadEmbeddedCommands()
}
