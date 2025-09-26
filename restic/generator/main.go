package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/creativeprojects/resticprofile/restic"
)

var (
	commandsFile string
	installDir   string
	version      string
	baseVersion  bool
	manualDir    string
)

func commandNames() []string { return restic.CommandNamesForVersion(restic.AnyVersion) }

func generate() (err error) {
	// Load existing commands (if any)
	restic.ClearCommands()
	_ = restic.LoadCommands(commandsFile)
	prevNames := commandNames()

	// Parse man pages
	err = restic.ParseCommandsFromManPages(os.DirFS(manualDir), version, baseVersion)
	if err == nil {
		names := commandNames()
		fmt.Printf("found %d new commands in restic %s\n", len(names)-len(prevNames), version)
		fmt.Printf("saving to: %s\n", commandsFile)
		genFile := commandsFile + ".gen"
		err = restic.StoreCommands(genFile)
		if err == nil {
			err = restic.LoadCommands(genFile)
		}

		if err == nil {
			if len(names) == 0 {
				err = fmt.Errorf("found no commands")
			} else if len(names) != len(commandNames()) {
				err = fmt.Errorf("found %d commands but serialized state decodes to %d", len(names), len(commandNames()))
			}
		}

		if err == nil {
			os.Remove(commandsFile)
			err = os.Rename(genFile, commandsFile)
		} else {
			os.Remove(genFile)
		}
	}
	return
}

func install(includeManual bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	executable := path.Join(installDir, restic.Executable)

	// Install if required
	if s, err := os.Stat(executable); os.IsNotExist(err) {
		fmt.Println("installing: " + executable)
		if err = restic.DownloadBinary(executable, version); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if s.IsDir() {
		return fmt.Errorf("%s is a directory", executable)
	}

	// Check version
	if actualVersion, err := restic.GetVersion(executable); err == nil {
		if version == "" {
			version = actualVersion
		} else if version != actualVersion {
			return fmt.Errorf("restic version is %s while %s is required", actualVersion, version)
		}
	} else {
		return err
	}

	// Install manual
	if includeManual {
		manPage := path.Join(manualDir, "restic.1")
		if _, err := os.Stat(manPage); os.IsNotExist(err) {
			fmt.Println("creating man pages: " + manualDir)
			cmd := exec.CommandContext(ctx, executable, "generate", "--man", manualDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("man pages failed: %w", err)
			}
		}
	}

	return nil
}

func validateDir(dir string, createIfMissing bool) string {
	var err error
	if dir, err = filepath.Abs(dir); err == nil {
		if stat, e := os.Stat(dir); e == nil {
			if !stat.IsDir() {
				err = fmt.Errorf("is not a directory")
			}
		} else if os.IsNotExist(e) && createIfMissing {
			fmt.Println("creating: " + dir)
			err = os.MkdirAll(installDir, 0775)
		} else {
			err = e
		}
	}
	if err != nil {
		log.Fatalf("invalid dir parameter (create %s) %s", strconv.FormatBool(createIfMissing), err.Error())
	}
	return dir
}

func main() {
	flag.StringVar(&manualDir, "man", "", "restic manual pages directory")
	flag.StringVar(&installDir, "install", "", "install restic and manual pages to directory")
	flag.StringVar(&version, "version", "", "restic version to install")
	flag.BoolVar(&baseVersion, "base-version", false,
		"toggles whether the specified version (and any earlier version) is not added to the commands file as min version.")
	flag.StringVar(&commandsFile, "commands", "", "command dictionary output file")

	flag.Parse()

	if len(installDir) > 0 && len(manualDir) == 0 {
		manualDir = installDir
	}

	if len(manualDir) == 0 || len(commandsFile) == 0 {
		flag.PrintDefaults()
		return
	}

	var err error

	// Check dirs
	if len(installDir) > 0 {
		installDir = validateDir(installDir, true)
	}
	if len(manualDir) > 0 {
		manualDir = validateDir(manualDir, false)
	}

	// Check output file
	if len(commandsFile) > 0 {
		if commandsFile, err = filepath.Abs(commandsFile); err == nil {
			dir := filepath.Dir(commandsFile)
			if stat, e := os.Stat(dir); e == nil {
				if !stat.IsDir() {
					err = fmt.Errorf("%s is not a directory", dir)
				}
			} else {
				err = e
			}
		}
		if err != nil {
			log.Fatalf("invalid path to output file: %s", err.Error())
			return
		}
	}

	// do install
	if len(installDir) > 0 {
		includeManual := len(commandsFile) > 0
		if err = install(includeManual); err != nil {
			log.Fatalf("failed installing restic in %s: %s", installDir, err.Error())
		}
	}

	// do commands generation
	if len(commandsFile) > 0 {
		if err = generate(); err != nil {
			log.Fatalf("failed generating restic commands dictionary: %s", err.Error())
		}
	}
}
