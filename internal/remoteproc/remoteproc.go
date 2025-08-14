package remoteproc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Arm-Debug/remoteproc-simulator/internal/dirwatcher"
)

type RemoteProc struct {
	deviceDir string
	watcher   *dirwatcher.DirWatcher
	state     State
	firmware  string
	stopChan  chan struct{}
}

const (
	firmwareFileName   = "firmware"
	stateFileName      = "state"
	deviceNameFileName = "name"
	initialState       = StateOffline
	initialFirmware    = ""
)

func New(
	sysRoot string,
	deviceIndex uint,
	deviceName string,
) (*RemoteProc, error) {
	deviceDirName := fmt.Sprintf("remoteproc%d", deviceIndex)
	deviceDir := filepath.Join(sysRoot, "class", "remoteproc", deviceDirName)

	r := &RemoteProc{
		deviceDir: deviceDir,
		firmware:  initialFirmware,
		state:     initialState,
		stopChan:  make(chan struct{}),
	}

	files := map[string]string{
		stateFileName:      r.state.String(),
		firmwareFileName:   r.firmware,
		deviceNameFileName: deviceName,
	}
	if err := r.bootstrapDirectory(files); err != nil {
		return nil, fmt.Errorf("failed to bootstrap remote processor: %w", err)
	}

	watcher, err := dirwatcher.New(deviceDir)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("failed to setup directory watcher: %w", err)
	}

	r.watcher = watcher

	go r.loop()

	log.Printf("RemoteProc initialized at %s", r.deviceDir)
	return r, nil
}

func (r *RemoteProc) Close() error {
	err := r.watcher.Close()
	close(r.stopChan)
	return err
}

func (r *RemoteProc) bootstrapDirectory(files map[string]string) error {
	err := os.MkdirAll(r.deviceDir, 0755)
	if err != nil {
		return err
	}
	for filename, content := range files {
		if err := r.writeFile(filename, content); err != nil {
			return err
		}
	}
	return nil
}

func (r *RemoteProc) loop() {
	for {
		select {
		case <-r.stopChan:
			log.Printf("RemoteProc shutting down")
			return
		case event, ok := <-r.watcher.Changes():
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

func (r *RemoteProc) handleStateChange(value string) {
	if isStateSelfInflicted(value) {
		return
	}

	log.Printf("State change request: %s -> %s", r.state, value)

	switch value {
	case "start":
		if r.state == StateRunning {
			log.Printf("RemoteProc is already running")
			return
		}

		if r.firmware == "" {
			log.Printf("Cannot start: no firmware specified")
			r.state = StateCrashed
			r.setState(StateCrashed)
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
			log.Printf("RemoteProc is already stopped")
			return
		}

		log.Printf("Stopping remoteproc")
		r.setState(StateOffline)

	default:
		log.Printf("Invalid state command: %s", value)
	}
}

func (r *RemoteProc) setState(state State) {
	r.state = state
	r.writeFile(stateFileName, state.String())
}

func (r *RemoteProc) handleFirmwareChange(value string) {
	if r.state == StateRunning {
		log.Printf("Cannot change firmware while RemoteProc is %s", r.state)
		r.writeFile(firmwareFileName, r.firmware)
		return
	}
	r.firmware = value
	log.Printf("Firmware set to %s", value)
}

func (r *RemoteProc) writeFile(filename, content string) error {
	path := filepath.Join(r.deviceDir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", filename, err)
	}
	return nil
}

func isStateSelfInflicted(value string) bool {
	switch value {
	case StateRunning.String(), StateCrashed.String(), StateOffline.String():
		return true
	}
	return false
}
