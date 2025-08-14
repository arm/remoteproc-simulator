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

	log.Printf("RemoteProcessor initialized at %s", r.deviceDir)
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
			log.Printf("RemoteProcessor %s stopping...", r.deviceDir)
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
			log.Printf("RemoteProc %s is already running", r.deviceDir)
			return
		}

		if r.firmware == "" {
			log.Printf("Cannot start %s: no firmware specified", r.deviceDir)
			r.state = StateCrashed
			r.setState(StateCrashed)
			return
		}

		log.Printf("Starting remoteproc %s with firmware %s", r.deviceDir, r.firmware)

		// Simulate firmware loading delay
		go func() {
			select {
			case <-time.After(100 * time.Millisecond):
				log.Printf("Firmware %s loaded successfully", r.firmware)
				r.setState(StateRunning)
			case <-r.stopChan:
				log.Printf("Firmware loading cancelled due to shutdown")
			}
		}()

	case "stop":
		if r.state == StateOffline {
			log.Printf("RemoteProc %s is already stopped", r.deviceDir)
			return
		}

		log.Printf("Stopping remoteproc %s", r.deviceDir)
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
		log.Printf("Cannot change firmware while %s is %s", r.deviceDir, r.state)
		r.writeFile(firmwareFileName, r.firmware)
		return
	}
	r.firmware = value
	log.Printf("Firmware for %s set to: %s", r.deviceDir, value)
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
