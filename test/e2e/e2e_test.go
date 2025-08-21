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

		runSimulator(t, "--root-dir", root)

		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		requireState(t, instanceDir, "offline")
	})

	t.Run("default firmware is empty", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root-dir", root)

		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(instanceDir, "firmware"), "")
	})

	t.Run("instance name can be overriden", func(t *testing.T) {
		root := t.TempDir()
		name := "fancy-device"

		runSimulator(t, "--root-dir", root, "--name", name)

		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")
		assertFileContent(t, filepath.Join(instanceDir, "name"), name)
	})

	t.Run("instance index can be overriden", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root-dir", root, "--index", "99")

		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc99")
		assert.DirExists(t, instanceDir)
	})

	t.Run("firmware directory is created", func(t *testing.T) {
		root := t.TempDir()

		runSimulator(t, "--root-dir", root, "--index", "99")

		firmwareDir := filepath.Join(root, "lib", "firmware")
		assert.DirExists(t, firmwareDir)
	})
}

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

	t.Run("firmware file must exist in /lib/firmware", func(t *testing.T) {
		root := t.TempDir()
		runSimulator(t, "--root-dir", root, "--index", "0")
		instanceDir := filepath.Join(root, "sys", "class", "remoteproc", "remoteproc0")

		loadFirmware(t, instanceDir, "some-firmware.elf")
		setRemoteprocState(t, instanceDir, "start")

		requireState(t, instanceDir, "offline")
	})
}

func assertFileContent(t assert.TestingT, path string, wantContent string) {
	gotContent, err := os.ReadFile(path)
	if assert.NoError(t, err) {
		assert.Equal(t, wantContent, string(gotContent))
	}
}

func requireState(t *testing.T, instanceDir string, wantState string) {
	const waitFor = 500 * time.Millisecond
	const tickEvery = 100 * time.Millisecond
	stateFilePath := filepath.Join(instanceDir, "state")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assertFileContent(c, stateFilePath, wantState)
	}, waitFor, tickEvery)
}

func createFirmwareFile(t *testing.T, root, firmwareName string) {
	firmwarePath := filepath.Join(root, "lib", "firmware", firmwareName)
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
