package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/ktr0731/go-fuzzyfinder"
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
	lock := new(sync.Mutex)
	files := make([]RestorableFile, 0)

	command := "ls"
	profile, cleanup, err := openProfile(cmdCtx.config, cmdCtx.request.profile)
	defer cleanup()
	if err != nil {
		return err
	}
	cmdCtx.profile = profile

	cmdCtx.request.arguments = append(cmdCtx.request.arguments, "latest", "--long")

	wrapper := newResticWrapper(&cmdCtx.Context)
	args := profile.GetCommandFlags(command)
	rCommand := wrapper.prepareCommand(command, args, true)
	cmd := exec.CommandContext(context.TODO(), rCommand.command, rCommand.args...)
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()

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
			_, err := fmt.Sscanf(scanner.Text(), "%s %s %s %d %s %s %s", &file.mode, &file.uid, &file.gid, &file.size, &datePart, &timePart, &file.path)
			if err != nil {
				continue
			}
			file.date, _ = time.Parse(time.DateTime, datePart+" "+timePart)
			lock.Lock()
			files = append(files, file)
			lock.Unlock()
			firstOne()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file list:", err)
		}
		close(finished)
	})
	defer wg.Wait()

	// wait for at least one file to come in before displaying fuzzy finder
	select {
	case <-gotOne:
	case <-finished:
		return errors.New("no file to restore")
	}

	selected, err := fuzzyfinder.FindMulti(&files, func(i int) string {
		if i < 0 {
			return "error"
		}
		return files[i].path
	},
		fuzzyfinder.WithHeader("ENTER: select one, TAB: select multiple, ESC: cancel"),
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 {
				return "N/A"
			}
			uid, gid := files[i].uid, files[i].gid
			usr, err := user.LookupId(uid)
			if err == nil {
				uid = usr.Username
			}
			grp, err := user.LookupGroupId(gid)
			if err == nil {
				gid = grp.Name
			}
			return fmt.Sprintf(preview,
				filepath.Dir(files[i].path),
				filepath.Base(files[i].path),
				files[i].mode,
				uid,
				gid,
				files[i].size,
				files[i].date.Format(time.DateTime),
			)
		}),
		fuzzyfinder.WithHotReloadLock(lock),
	)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			clog.Info("operation cancelled")
			return nil
		}
		return err
	}
	fmt.Printf("%+v", selected)
	return nil
}
