package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
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
	files := make([]RestorableFile, 0, 100)
	go func() {
		for i := range 100 {
			lock.Lock()
			files = append(files, RestorableFile{path: fmt.Sprintf("file %d", i)})
			lock.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}()
	time.Sleep(20 * time.Millisecond)
	selected, err := fuzzyfinder.FindMulti(&files, func(i int) string {
		return files[i].path
	},
		fuzzyfinder.WithHeader("ENTER: select one, TAB: select multiple, ESC: cancel"),
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			return fmt.Sprintf("mode: %s\npath: %s", files[i].mode, files[i].path)
		}),
		fuzzyfinder.WithHotReloadLock(lock),
	)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", selected)
	return nil
}
