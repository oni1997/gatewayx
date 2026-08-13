package main

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watchConfigFile(path string, onReload func()) func() {
	if path == "" {
		return func() {}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return func() {}
	}

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		return func() {}
	}

	done := make(chan struct{})

	go func() {
		debounce := time.NewTimer(time.Hour)
		debounce.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Clean(event.Name) != filepath.Clean(path) {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					debounce.Reset(500 * time.Millisecond)
				}
			case <-debounce.C:
				slog.Info("config file changed, reloading...")
				onReload()
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
		_ = watcher.Close()
	}
}
