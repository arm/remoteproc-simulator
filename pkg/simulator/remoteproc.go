package simulator

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/arm/remoteproc-simulator/internal/dirwatcher"
)

type Remoteproc struct {
	name     string
	fs       *FileSystemManager
	watcher  *dirwatcher.DirWatcher
	state    state
	firmware string
	stopChan chan struct{}
}

const (
	firmwareFileName = "firmware"
	stateFileName    = "state"
	nameFileName     = "name"
	initialState     = StateOffline
	initialFirmware  = ""
)

type Config struct {
	// RootDir is the location where /sys and /lib will be created
	RootDir string
	// Index is the N in /sys/class/remoteproc/remoteprocN/.../
	Index uint
	// Name is the remote processor name written to /sys/class/remoteproc/.../name
	Name string
}

func (c Config) validate() error {
	if c.RootDir == "" {
		return errors.New("root directory must be specified")
	}
	if c.Name == "" {
		return errors.New("name name must be specified")
	}
	return nil
}

// NewRemoteproc creates a new [Remoteproc].
// The caller should call Close when finished to clean up resources.
func NewRemoteproc(config Config) (*Remoteproc, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	r := &Remoteproc{
		name:     config.Name,
		fs:       NewFileSystemManager(config.RootDir, config.Index),
		firmware: initialFirmware,
		state:    initialState,
	}

	err := r.start()
	if err != nil {
		r.Close()
	}
	return r, err
}

func (r *Remoteproc) start() error {
	if err := r.bootstrapDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to bootstrap directory structure: %w", err)
	}

	watcher, err := dirwatcher.New(r.fs.InstanceDir())
	if err != nil {
		return fmt.Errorf("failed to setup directory watcher: %w", err)
	}
	r.watcher = watcher

	r.stopChan = make(chan struct{})
	go r.loop()

	log.Printf("Remoteproc initialized at %s", r.fs.InstanceDir())
	return nil
}

func (r *Remoteproc) Close() error {
	var watcherErr error
	if r.watcher != nil {
		watcherErr = r.watcher.Close()
	}

	if r.stopChan != nil {
		close(r.stopChan)
	}

	var fsErr error
	if r.fs != nil {
		fsErr = r.fs.Cleanup()
	}

	return errors.Join(watcherErr, fsErr)
}

func (r *Remoteproc) bootstrapDirectoryStructure() error {
	if err := r.fs.BootstrapDirectories(); err != nil {
		return err
	}

	files := map[string]string{
		stateFileName:    r.state.String(),
		firmwareFileName: r.firmware,
		nameFileName:     r.name,
	}

	for filename, content := range files {
		if err := r.fs.WriteInstanceFile(filename, content); err != nil {
			return err
		}
	}

	return nil
}

func (r *Remoteproc) loop() {
	for {
		select {
		case <-r.stopChan:
			log.Printf("Remoteproc shutting down")
			return
		case event, ok := <-r.watcher.Changes():
			log.Printf("Remoteproc detected change: %s = %s", event.Filename, event.Value)
			if !ok {
				return
			}
			switch event.Filename {
			case "state":
				r.handleStateChange(event.Value)
			case firmwareFileName:
				r.handleFirmwareChange(event.Value)
			}
		}
	}
}

func (r *Remoteproc) handleStateChange(value string) {
	if isStateSelfInflicted(value) {
		return
	}

	log.Printf("State change request: %s -> %s", r.state, value)

	switch value {
	case "start":
		if r.state == StateRunning {
			log.Printf("Remoteproc is already running")
			return
		}

		if r.firmware == "" {
			log.Printf("Cannot start: no firmware specified")
			r.state = StateCrashed
			r.setState(StateCrashed)
			return
		}

		if err := r.fs.CheckFirmwareExists(r.firmware); err != nil {
			log.Printf("Cannot start: %s", err)
			r.setState(r.state)
			return
		}

		log.Printf("Starting remoteproc with firmware %s", r.firmware)

		// Simulate firmware loading delay
		go func() {
			select {
			case <-time.After(100 * time.Millisecond):
				log.Printf("Firmware %s started successfully", r.firmware)
				r.setState(StateRunning)
			case <-r.stopChan:
				log.Printf("Firmware loading cancelled due to shutdown")
			}
		}()

	case "stop":
		if r.state == StateOffline {
			log.Printf("Remoteproc is already stopped")
			return
		}

		log.Printf("Stopping remoteproc")
		r.setState(StateOffline)

	default:
		log.Printf("Invalid state command: %s", value)
	}
}

func (r *Remoteproc) setState(state state) {
	r.state = state
	r.fs.WriteInstanceFile(stateFileName, state.String())
}

func (r *Remoteproc) handleFirmwareChange(value string) {
	if r.state == StateRunning {
		log.Printf("Cannot change firmware while Remoteproc is %s", r.state)
		r.fs.WriteInstanceFile(firmwareFileName, r.firmware)
		return
	}
	r.firmware = value
	log.Printf("Firmware set to %s", value)
}

func isStateSelfInflicted(value string) bool {
	switch value {
	case StateRunning.String(), StateCrashed.String(), StateOffline.String():
		return true
	}
	return false
}
