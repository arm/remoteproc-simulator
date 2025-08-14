package e2e

import (
	"bufio"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func runSimulator(t *testing.T, args ...string) {
	t.Helper()
	bin := buildSimulatorBin(t)

	simulatorCmd := exec.Command(bin, args...)

	stderr, err := simulatorCmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	if err := simulatorCmd.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}

	t.Cleanup(func() {
		if simulatorCmd.Process != nil {
			simulatorCmd.Process.Kill()
		}
	})

	requireReady(t, stderr)
}

func buildSimulatorBin(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "remoteproc-simulator")
	cmd := exec.Command("go", "build", "-o", outputPath, "../../cmd/remoteproc-simulator")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	return outputPath
}

func requireReady(t *testing.T, output io.Reader) {
	t.Helper()
	const waitFor = 1000 * time.Millisecond
	const tickEvery = 100 * time.Millisecond
	scanner := bufio.NewScanner(output)
	require.Eventually(t, func() bool {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "RemoteProcessor initialized") {
				return true
			}
		}
		return false
	}, waitFor, tickEvery, "simulator not ready within timeout")
}
