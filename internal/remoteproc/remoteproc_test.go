package remoteproc_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Arm-Debug/remoteproc-simulator/internal/remoteproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootsrappingDeviceDir(t *testing.T) {
	sysRoot := t.TempDir()
	var deviceIndex uint = 1
	deviceName := "fancy-device"

	r, err := remoteproc.New(sysRoot, deviceIndex, deviceName)
	defer func() { _ = r.Close() }()
	require.NoError(t, err)

	deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc1")
	assertFileContent(t, filepath.Join(deviceDir, "state"), "offline")
	assertFileContent(t, filepath.Join(deviceDir, "firmware"), "")
	assertFileContent(t, filepath.Join(deviceDir, "name"), deviceName)
}

func TestStartingAndStoppingFirmware(t *testing.T) {
	sysRoot := t.TempDir()
	var deviceIndex uint = 0
	r, err := remoteproc.New(sysRoot, deviceIndex, "dont-care")
	defer func() { _ = r.Close() }()
	require.NoError(t, err)
	deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc0")
	stateFilePath := filepath.Join(deviceDir, "state")

	// Load firmware and start remoteproc
	require.NoError(t, writeFile(filepath.Join(deviceDir, "firmware"), "some-firmware.elf"))
	require.NoError(t, writeFile(stateFilePath, "start"))

	assertState(t, deviceDir, "running")

	// Stop remoteproc
	require.NoError(t, writeFile(stateFilePath, "stop"))

	assertState(t, deviceDir, "offline")
}

func TestCantStartWithoutFirmware(t *testing.T) {
	sysRoot := t.TempDir()
	var deviceIndex uint = 0
	r, err := remoteproc.New(sysRoot, deviceIndex, "dont-care")
	defer func() { _ = r.Close() }()
	require.NoError(t, err)
	deviceDir := filepath.Join(sysRoot, "class", "remoteproc", "remoteproc0")
	stateFilePath := filepath.Join(deviceDir, "state")

	require.NoError(t, writeFile(stateFilePath, "start"))

	assertState(t, deviceDir, "crashed")
}

func assertState(t *testing.T, deviceDir string, wantState string) {
	const waitFor = 500 * time.Millisecond
	const tickEvery = 100 * time.Millisecond
	stateFilePath := filepath.Join(deviceDir, "state")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assertFileContent(c, stateFilePath, wantState)
	}, waitFor, tickEvery)
}

func assertFileContent(t assert.TestingT, path string, wantContent string) {
	gotContent, err := os.ReadFile(path)
	if assert.NoError(t, err) {
		assert.Equal(t, wantContent, string(gotContent))
	}
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
