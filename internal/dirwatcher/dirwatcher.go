package dirwatcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type FileChangeEvent struct {
	Filename string
	Value    string
}

type DirWatcher struct {
	watcher      *fsnotify.Watcher
	changeEvents chan FileChangeEvent
}

func New(path string) (*DirWatcher, error) {
	watcher, err := setupFSNotify(path)
	if err != nil {
		return nil, err
	}
	d := &DirWatcher{
		watcher:      watcher,
		changeEvents: make(chan FileChangeEvent),
	}
	go d.loop()
	return d, nil
}

func (d *DirWatcher) Changes() <-chan FileChangeEvent {
	return d.changeEvents
}

func (d *DirWatcher) Close() error {
	return d.watcher.Close()
}

func (d *DirWatcher) loop() {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				close(d.changeEvents)
				return
			}

			if event.Op&fsnotify.Write == 0 {
				continue
			}

			filename := filepath.Base(event.Name)
			content, err := os.ReadFile(event.Name)
			if err != nil {
				log.Printf("Error reading %s: %v", filename, err)
				continue
			}
			value := strings.TrimSpace(string(content))
			d.changeEvents <- FileChangeEvent{
				Filename: filename,
				Value:    value,
			}

		case err, ok := <-d.watcher.Errors:
			if !ok {
				close(d.changeEvents)
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func setupFSNotify(path string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch sysfs directory: %v", err)
	}

	return watcher, nil
}
