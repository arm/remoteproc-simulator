package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapping(t *testing.T) {
	t.Run("default state is offline", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root", root)

		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		requireState(t, deviceDir, "offline")
	})

	t.Run("default firmware is empty", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root", root)

		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(deviceDir, "firmware"), "")
	})

	t.Run("device name can be overriden", func(t *testing.T) {
		root := t.TempDir()
		deviceName := "fancy-device"

		runSimulator(t, "--root", root, "--device-name", deviceName)

		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(deviceDir, "name"), deviceName)
	})

	t.Run("device index can be overriden", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root", root, "--device-index", "99")

		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc99")
		assert.DirExists(t, deviceDir)
	})

	t.Run("firmware directory is created", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root", root, "--device-index", "99")

		firmwareDir := filepath.Join(root, "lib", "firmware")
		assert.DirExists(t, firmwareDir)
	})
}

func TestRunningFirmware(t *testing.T) {
	t.Run("starting firmware sets state to running, stopping to offline", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root", root, "--device-index", "3")
		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc3")

		createFirmwareFile(t, root, "some-firmware.elf")
		loadFirmware(t, deviceDir, "some-firmware.elf")
		setRemoteprocState(t, deviceDir, "start")

		requireState(t, deviceDir, "running")

		setRemoteprocState(t, deviceDir, "stop")

		requireState(t, deviceDir, "offline")
	})

	t.Run("firmware file must exist in /lib/firmware", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root", root, "--device-index", "0")
		deviceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

		loadFirmware(t, deviceDir, "some-firmware.elf")
		setRemoteprocState(t, deviceDir, "start")

		requireState(t, deviceDir, "offline")
	})
}

func assertFileContent(t assert.TestingT, path string, wantContent string) {
	gotContent, err := os.ReadFile(path)
	if assert.NoError(t, err) {
		assert.Equal(t, wantContent, string(gotContent))
	}
}

func requireState(t *testing.T, deviceDir string, wantState string) {
	const waitFor = 500 * time.Millisecond
	const tickEvery = 100 * time.Millisecond
	stateFilePath := filepath.Join(deviceDir, "state")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assertFileContent(c, stateFilePath, wantState)
	}, waitFor, tickEvery)
}

func createFirmwareFile(t *testing.T, root, firmwareName string) {
	firmwarePath := filepath.Join(root, "lib", "firmware", firmwareName)
	require.NoError(t, writeFile(firmwarePath, ""))
}

func loadFirmware(t *testing.T, deviceDir, firmwareName string) {
	firmwareFilePath := filepath.Join(deviceDir, "firmware")
	require.NoError(t, writeFile(firmwareFilePath, firmwareName))
}

func setRemoteprocState(t *testing.T, deviceDir, state string) {
	stateFilePath := filepath.Join(deviceDir, "state")
	require.NoError(t, writeFile(stateFilePath, state))
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
