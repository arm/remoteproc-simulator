package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arm-Debug/remoteproc-simulator/internal/remoteproc"
	"github.com/spf13/cobra"
)

func main() {
	var sysfsRoot string
	var deviceIndex uint
	var deviceName string

	rootCmd := &cobra.Command{
		Use:   "remoteproc-simulator",
		Short: "RemoteProc Simulator - Linux remoteproc subsystem simulator",
		Long: `RemoteProc Simulator simulates the Linux remoteproc subsystem for testing purposes.

Example usage:
  # Start daemon with custom options
  remoteproc-simulator --sysfs /tmp/fake-root/sys --device-index 0 --device-name dsp0

  # In another terminal, control via sysfs:
  echo 'hello_world.fw' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware
  echo 'start' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/name  # Shows 'dsp0'
  echo 'stop' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state`,
		Run: func(cmd *cobra.Command, args []string) {
			remoteProcessor, err := remoteproc.New(sysfsRoot, deviceIndex, deviceName)
			if err != nil {
				log.Fatalf("Failed to create remote processor: %v", err)
			}
			defer remoteProcessor.Close()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			<-sigChan
			log.Println("Received shutdown signal")
		},
	}

	rootCmd.Flags().UintVar(&deviceIndex, "device-index", 0, "Device index (suffix of device directory)")
	rootCmd.Flags().StringVar(&deviceName, "device-name", "dsp0", "Device name identifier (appears in the 'name' sysfs file)")
	rootCmd.Flags().StringVar(&sysfsRoot, "sysfs", "/tmp/fake-root/sys", "Sysfs root path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
