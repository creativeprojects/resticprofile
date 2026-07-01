package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	fzf "github.com/junegunn/fzf/src"
)

const (
	preview = `
path: %s
file: %s
mode: %s
user: %s
grp:  %s
size: %d
time: %s
`
)

type RestorableFile struct {
	mode string
	uid  string
	gid  string
	size int64
	date time.Time
	path string
}

func fuzzyRestore(cmdCtx commandContext) error {
	// lock := new(sync.Mutex)
	// files := make([]RestorableFile, 0)
	filesChan := make(chan string, 100)

	// command := "ls"
	// profile, cleanup, err := openProfile(cmdCtx.config, cmdCtx.request.profile)
	// defer cleanup()
	// if err != nil {
	// 	return err
	// }
	// cmdCtx.profile = profile

	// cmdCtx.request.arguments = append(cmdCtx.request.arguments, "latest", "--long")

	// wrapper := newResticWrapper(&cmdCtx.Context)
	// args := profile.GetCommandFlags(command)
	// rCommand := wrapper.prepareCommand(command, args, true)
	// shellCmd := shell.NewCommand(rCommand.command, rCommand.args)
	// shellCmd.Shell = rCommand.shell
	// shellCmd.Stderr = os.Stderr
	// shellCmd.Dir = rCommand.dir

	// reader, writer := io.Pipe()
	// shellCmd.Stdout = writer

	// shellCommand, shellArgs, err := shellCmd.GetShellCommand()
	// if err != nil {
	// 	return err
	// }
	// clog.Warning(shellCommand, shellArgs)
	// cmd := exec.CommandContext(context.TODO(), shellCommand, shellArgs...)
	// if err = cmd.Start(); err != nil {
	// 	return err
	// }
	// defer cmd.Wait()

	reader, err := os.Open("output.txt")
	if err != nil {
		return err
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	gotOne := make(chan struct{})
	finished := make(chan struct{})
	firstOne := sync.OnceFunc(func() {
		close(gotOne)
	})
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		for scanner.Scan() {
			file := RestorableFile{}
			var datePart, timePart string
			line := scanner.Text()
			_, err := fmt.Sscanf(line, "%s %s %s %d %s %s %s", &file.mode, &file.uid, &file.gid, &file.size, &datePart, &timePart, &file.path)
			if err != nil {
				continue
			}
			filesChan <- line
			// file.date, _ = time.Parse(time.DateTime, datePart+" "+timePart)
			// lock.Lock()
			// files = append(files, file)
			// lock.Unlock()
			firstOne()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file list:", err)
		}
		close(finished)
		close(filesChan)
	})
	defer wg.Wait()

	// wait for at least one file to come in before displaying fuzzy finder
	select {
	case <-gotOne:
	case <-finished:
		return errors.New("no file to restore")
	}

	// selected, err := fuzzyfinder.FindMulti(&files, func(i int) string {
	// 	if i < 0 {
	// 		return "error"
	// 	}
	// 	return files[i].path
	// },
	// 	fuzzyfinder.WithHeader("ENTER: select one, TAB: select multiple, ESC: cancel"),
	// 	fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
	// 		if i < 0 {
	// 			return "N/A"
	// 		}
	// 		uid, gid := files[i].uid, files[i].gid
	// 		usr, err := user.LookupId(uid)
	// 		if err == nil {
	// 			uid = usr.Username
	// 		}
	// 		grp, err := user.LookupGroupId(gid)
	// 		if err == nil {
	// 			gid = grp.Name
	// 		}
	// 		return fmt.Sprintf(preview,
	// 			filepath.Dir(files[i].path),
	// 			filepath.Base(files[i].path),
	// 			files[i].mode,
	// 			uid,
	// 			gid,
	// 			files[i].size,
	// 			files[i].date.Format(time.DateTime),
	// 		)
	// 	}),
	// 	fuzzyfinder.WithHotReloadLock(lock),
	// )
	// if err != nil {
	// 	if errors.Is(err, fuzzyfinder.ErrAbort) {
	// 		clog.Info("operation cancelled")
	// 		return nil
	// 	}
	// 	return err
	// }
	// fmt.Printf("%+v", selected)
	options, err := fzf.ParseOptions(
		true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
		[]string{
			"--multi",
			"--scheme", "path",
			"--with-nth", "-1",
			"--accept-nth", "-1",
			"--preview", "echo {1} {2} {3} {4} {5} {6}; dirname {-1}; basename {-1}",
			"--preview-window", "down,3",
		},
	)
	if err != nil {
		return err
	}

	selectedChan := make(chan string, 1)
	wg.Go(func() {
		for selectedFile := range selectedChan {
			fmt.Println(selectedFile)
		}
	})
	options.Input = filesChan
	options.Output = selectedChan

	// Run fzf
	_, err = fzf.Run(options)
	close(selectedChan)
	if err != nil {
		return err
	}
	return nil
}
