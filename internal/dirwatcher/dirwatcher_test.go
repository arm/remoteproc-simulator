package dirwatcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Arm-Debug/remoteproc-simulator/internal/dirwatcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirWatcher(t *testing.T) {
	t.Run("it notifies when file contents change in a watched directory", func(t *testing.T) {
		watchedDir := t.TempDir()
		os.WriteFile(filepath.Join(watchedDir, "foo.txt"), []byte(""), 0644)
		os.WriteFile(filepath.Join(watchedDir, "bar.txt"), []byte(""), 0644)

		watcher, err := dirwatcher.New(watchedDir)
		require.NoError(t, err)

		os.WriteFile(filepath.Join(watchedDir, "foo.txt"), []byte("hi"), 0644)
		os.WriteFile(filepath.Join(watchedDir, "bar.txt"), []byte("aloha"), 0644)

		want := []dirwatcher.FileChangeEvent{
			{Filename: "foo.txt", Value: "hi"},
			{Filename: "bar.txt", Value: "aloha"},
		}

		var got []dirwatcher.FileChangeEvent

		waitFor := time.After(100 * time.Millisecond)
		for len(got) < len(want) {
			select {
			case event, ok := <-watcher.Changes():
				if !ok {
					break
				}
				got = append(got, event)
			case <-waitFor:
				t.Fatalf("Did not receive %d messages, got %d: %#v", len(want), len(got), got)
			}
		}

		assert.Equal(t, want, got)
	})
}
