package e2e

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runSimulator(t *testing.T, args ...string) {
	t.Helper()
	bin := buildSimulatorBin(t)

	simulatorCmd := exec.Command(bin, args...)

	stderr, err := simulatorCmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	if err := simulatorCmd.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}

	var stderrBuf bytes.Buffer
	stderrTee := io.TeeReader(stderr, &stderrBuf)
	readyCh := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stderrTee)
		isReady := false
		for scanner.Scan() {
			line := scanner.Text()
			if !isReady {
				if strings.Contains(line, "RemoteProc initialized") {
					isReady = true
					close(readyCh)
				}
			}
		}
	}()

	t.Cleanup(func() {
		if simulatorCmd.Process != nil {
			simulatorCmd.Process.Kill()
		}
		if t.Failed() {
			t.Logf("Simulator output:\n%s", stderrBuf.String())
		}
	})

	select {
	case <-readyCh:
		return
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Simulator not ready within timeout")
	}
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
