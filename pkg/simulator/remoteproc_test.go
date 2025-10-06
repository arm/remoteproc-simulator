package simulator_test

import (
	"testing"

	"github.com/arm/remoteproc-simulator/pkg/simulator"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	t.Run("it requires root directory to be specified", func(t *testing.T) {
		invalidConfig := simulator.Config{RootDir: ""}

		_, err := simulator.NewRemoteproc(invalidConfig)

		assert.ErrorContains(t, err, "root directory must be specified")
	})

	t.Run("it requires name to be specified", func(t *testing.T) {
		invalidConfig := simulator.Config{RootDir: "some/dir", Name: ""}

		_, err := simulator.NewRemoteproc(invalidConfig)

		assert.ErrorContains(t, err, "name must be specified")
	})
}
