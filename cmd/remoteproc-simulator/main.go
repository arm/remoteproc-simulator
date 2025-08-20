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

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var rootDir string
	var deviceIndex uint
	var deviceName string
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:   "remoteproc-simulator",
		Short: "Remoteproc Simulator - Linux remoteproc subsystem simulator",
		Long: `Remoteproc Simulator simulates the Linux remoteproc subsystem for testing purposes.

Example usage:
  # Start daemon with custom options
  remoteproc-simulator --root-dir /tmp/fake-root --device-index 0 --device-name dsp0

  # In another terminal, control via sysfs:
  touch /tmp/fake-root/lib/firmware/hello_world.elf
  echo hello_world.elf > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware
  echo start > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/name  # Shows 'dsp0'
  echo stop > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state`,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Println("remoteproc-simulator")
				fmt.Printf("  version: %s\n", version)
				if commit != "none" {
					fmt.Printf("  commit: %s\n", commit)
				}
				if date != "unknown" {
					fmt.Printf("  built at: %s\n", date)
				}
				os.Exit(0)
			}
			remoteProcessor, err := remoteproc.New(rootDir, deviceIndex, deviceName)
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

	rootCmd.Flags().UintVar(&deviceIndex, "device-index", 0, "suffix of device directory (default 0)")
	rootCmd.Flags().StringVar(&deviceName, "device-name", "dsp0", "device name identifier")
	rootCmd.Flags().StringVar(&rootDir, "root-dir", "/tmp/fake-root", "directory where /sys and /lib will be created")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "show version information")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
