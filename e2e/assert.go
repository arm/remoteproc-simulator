package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
