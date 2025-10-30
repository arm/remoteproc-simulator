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

		createFirmwareFile(t, filepath.Join(root, "lib", "firmware", "some-firmware.elf"))
		loadFirmware(t, instanceDir, "some-firmware.elf")
		setRemoteprocState(t, instanceDir, "start")

		requireState(t, instanceDir, "running")

		setRemoteprocState(t, instanceDir, "stop")

		requireState(t, instanceDir, "offline")
	})

	t.Run("firmware file must exist in order to start remoteproc successfully", func(t *testing.T) {
		t.Run("in default firmware directory - /lib/firmware", func(t *testing.T) {
			root := t.TempDir()
			runSimulator(t, "--root-dir", root, "--index", "0")
			instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

			createFirmwareFile(t, filepath.Join(root, "lib", "firmware", "some-firmware.elf"))
			loadFirmware(t, instanceDir, "some-firmware.elf")
			setRemoteprocState(t, instanceDir, "start")

			requireState(t, instanceDir, "running")
		})

		t.Run("in custom firmware directory, specified in /sys/module/firmware_class/parameters/path", func(t *testing.T) {
			root := t.TempDir()
			runSimulator(t, "--root-dir", root, "--index", "0")
			instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
			customFirmwarePath := t.TempDir()
			firmwareSearchPathFile := filepath.Join(root, "sys", "module", "firmware_class", "parameters", "path")

			require.NoError(t, writeFile(firmwareSearchPathFile, customFirmwarePath))
			createFirmwareFile(t, filepath.Join(customFirmwarePath, "some-firmware.elf"))
			loadFirmware(t, instanceDir, "some-firmware.elf")
			setRemoteprocState(t, instanceDir, "start")

			requireState(t, instanceDir, "running")
		})

		t.Run("when firmware file doesn't exist, remoteproc fails to start", func(t *testing.T) {
			root := t.TempDir()
			runSimulator(t, "--root-dir", root, "--index", "0")
			instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

			loadFirmware(t, instanceDir, "some-firmware.elf")
			setRemoteprocState(t, instanceDir, "start")

			requireState(t, instanceDir, "offline")
		})
	})

	t.Run("firmware file must exist in folder indicated inside <fake-root>/sys/module/firmware_class/parameters/path", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root-dir", root, "--index", "0")
		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

		loadFirmware(t, instanceDir, "some-firmware.elf")
		setRemoteprocState(t, instanceDir, "start")

		requireState(t, instanceDir, "offline")
	})
}

func createFirmwareFile(t *testing.T, pathToFirmwareFile string) {
	firmwarePath := filepath.Join(pathToFirmwareFile)
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
