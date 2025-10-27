package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunningFirmware(t *testing.T) {
	t.Run("starting firmware sets state to running, stopping to offline", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root-dir", root, "--index", "3")
		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc3")

		createFirmwareFile(t, root, "some-firmware.elf")
		loadFirmware(t, instanceDir, "some-firmware.elf")
		setRemoteprocState(t, instanceDir, "start")

		requireState(t, instanceDir, "running")

		setRemoteprocState(t, instanceDir, "stop")

		requireState(t, instanceDir, "offline")
	})

	t.Run("firmware file must exist in folder indicated inside /sys/module/firmware_class/parameters/path", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root-dir", root, "--index", "0")
		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

		loadFirmware(t, instanceDir, "some-firmware.elf")
		setRemoteprocState(t, instanceDir, "start")

		requireState(t, instanceDir, "offline")
	})
}

func createFirmwareFile(t *testing.T, root, firmwareName string) {
	firmwareDirPath, err := os.ReadFile(filepath.Join(root, "sys", "module", "firmware_class", "parameters", "path"))
	require.NoError(t, err)
	os.MkdirAll(string(firmwareDirPath), 0755)
	firmwarePath := filepath.Join(string(firmwareDirPath), firmwareName)
	require.NoError(t, writeFile(firmwarePath, ""))
}

func loadFirmware(t *testing.T, instanceDir, firmwareName string) {
	firmwareFilePath := filepath.Join(instanceDir, "firmware")
	require.NoError(t, writeFile(firmwareFilePath, firmwareName))
}

func setRemoteprocState(t *testing.T, instanceDir, state string) {
	stateFilePath := filepath.Join(instanceDir, "state")
	require.NoError(t, writeFile(stateFilePath, state))
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
