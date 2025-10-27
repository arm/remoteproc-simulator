package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

		firmwareDir, err := os.ReadFile(filepath.Join(root, "sys", "module", "firmware_class", "parameters", "path"))
		assert.NoError(t, err)
		assert.DirExists(t, string(firmwareDir))
	})
}
