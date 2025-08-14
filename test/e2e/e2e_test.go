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
		sysRoot := t.TempDir()

		runSimulator(t, "--sysfs", sysRoot)

		deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(deviceDir, "state"), "offline")
	})

	t.Run("default firmware is empty", func(t *testing.T) {
		sysRoot := t.TempDir()

		runSimulator(t, "--sysfs", sysRoot)

		deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(deviceDir, "firmware"), "")
	})

	t.Run("device name can be overriden", func(t *testing.T) {
		sysRoot := t.TempDir()
		deviceName := "fancy-device"

		runSimulator(t, "--sysfs", sysRoot, "--device-name", deviceName)

		deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(deviceDir, "name"), deviceName)
	})

	t.Run("device index can be overriden", func(t *testing.T) {
		sysRoot := t.TempDir()

		runSimulator(t, "--sysfs", sysRoot, "--device-index", "99")

		deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc99")
		assert.DirExists(t, deviceDir)
	})
}

func TestStartAndStop(t *testing.T) {
	t.Run("starting firmware sets state to running, stopping to offline", func(t *testing.T) {
		sysRoot := t.TempDir()
		runSimulator(t, "--sysfs", sysRoot, "--device-index", "3")
		deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc3")
		stateFilePath := filepath.Join(deviceDir, "state")

		// Load firmware and start remoteproc
		require.NoError(t, writeFile(filepath.Join(deviceDir, "firmware"), "some-firmware.elf"))
		require.NoError(t, writeFile(stateFilePath, "start"))

		requireState(t, deviceDir, "running")

		// Stop remoteproc
		require.NoError(t, writeFile(stateFilePath, "stop"))

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

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
