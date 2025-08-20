package simulator

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Arm-Debug/remoteproc-simulator/internal/dirwatcher"
)

type Remoteproc struct {
	rootDir     string
	deviceIndex uint
	deviceName  string

	deviceDir   string
	firmwareDir string
	watcher     *dirwatcher.DirWatcher
	state       state
	firmware    string
	stopChan    chan struct{}
}

const (
	firmwareFileName   = "firmware"
	stateFileName      = "state"
	deviceNameFileName = "name"
	initialState       = StateOffline
	initialFirmware    = ""
)

type Config struct {
	// RootDir is the location where /sys and /lib will be created
	RootDir string
	// DeviceIndex is the N in remoteprocN device directory
	DeviceIndex uint
	// DeviceName is the device name identifier written to /sys/class/remoteproc/.../name
	DeviceName string
}

func (c Config) validate() error {
	if c.RootDir == "" {
		return errors.New("root directory must be specified")
	}
	if c.DeviceName == "" {
		return errors.New("device name must be specified")
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
		rootDir:     config.RootDir,
		deviceIndex: config.DeviceIndex,
		deviceName:  config.DeviceName,
		firmware:    initialFirmware,
		state:       initialState,
	}

	deviceDirName := fmt.Sprintf("remoteproc%d", r.deviceIndex)
	r.deviceDir = filepath.Join(r.rootDir, "sys", "class", "remoteproc", deviceDirName)
	r.firmwareDir = filepath.Join(r.rootDir, "lib", "firmware")

	return r, r.start()
}

func (r *Remoteproc) start() error {
	files := map[string]string{
		stateFileName:      r.state.String(),
		firmwareFileName:   r.firmware,
		deviceNameFileName: r.deviceName,
	}
	if err := r.bootstrapDeviceDir(files); err != nil {
		return fmt.Errorf("failed to bootstrap sysfs: %w", err)
	}

	if err := r.bootstrapFirmwareDir(); err != nil {
		return fmt.Errorf("failed to bootstrap firmware dir: %w", err)
	}

	watcher, err := dirwatcher.New(r.deviceDir)
	if err != nil {
		return fmt.Errorf("failed to setup directory watcher: %w", err)
	}

	r.watcher = watcher

	r.stopChan = make(chan struct{})
	go r.loop()

	log.Printf("Remoteproc initialized at %s", r.deviceDir)
	return nil
}

func (r *Remoteproc) Close() error {
	var err error
	if r.watcher != nil {
		err = r.watcher.Close()
	}
	if r.stopChan != nil {
		close(r.stopChan)
	}
	return err
}

func (r *Remoteproc) bootstrapDeviceDir(files map[string]string) error {
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

func (r *Remoteproc) bootstrapFirmwareDir() error {
	return os.MkdirAll(r.firmwareDir, 0755)
}

func (r *Remoteproc) loop() {
	for {
		select {
		case <-r.stopChan:
			log.Printf("Remoteproc shutting down")
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

		if !fileExists(filepath.Join(r.firmwareDir, r.firmware)) {
			log.Printf("Cannot start: firmware file not found in %s directory", r.firmwareDir)
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
	r.writeFile(stateFileName, state.String())
}

func (r *Remoteproc) handleFirmwareChange(value string) {
	if r.state == StateRunning {
		log.Printf("Cannot change firmware while Remoteproc is %s", r.state)
		r.writeFile(firmwareFileName, r.firmware)
		return
	}
	r.firmware = value
	log.Printf("Firmware set to %s", value)
}

func (r *Remoteproc) writeFile(filename, content string) error {
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
