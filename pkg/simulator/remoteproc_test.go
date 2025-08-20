package simulator_test

import (
	"testing"

	"github.com/Arm-Debug/remoteproc-simulator/pkg/simulator"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	t.Run("it requires root directory to be specified", func(t *testing.T) {
		invalidConfig := simulator.Config{RootDir: ""}

		_, err := simulator.NewRemoteproc(invalidConfig)

		assert.ErrorContains(t, err, "root directory must be specified")
	})

	t.Run("it requires device name to be specified", func(t *testing.T) {
		invalidConfig := simulator.Config{RootDir: "some/dir", DeviceName: ""}

		_, err := simulator.NewRemoteproc(invalidConfig)

		assert.ErrorContains(t, err, "device name must be specified")
	})
}
