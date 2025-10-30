package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arm/remoteproc-simulator/pkg/simulator"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var rootDir string
	var index uint
	var name string
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:   "remoteproc-simulator",
		Short: "Remoteproc Simulator - Linux remoteproc subsystem simulator",
		Long: `Remoteproc Simulator simulates the Linux remoteproc subsystem for testing purposes.

Example usage:
  # Start daemon with custom options
  remoteproc-simulator --root-dir /tmp/fake-root --index 0 --name dsp0

  # In another terminal, control via sysfs:
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/name  # Shows 'dsp0'

  touch /tmp/fake-root/lib/firmware/hello_world.elf
  echo hello_world.elf > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware

  echo start > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state  # Shows 'running'

  echo stop > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
  cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state  # Shows 'offline'
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if !cmd.Flags().Changed("root-dir") {
				tmpDir, err := os.MkdirTemp("", "remoteproc-simulator-*")
				if err != nil {
					return err
				}
				rootDir = tmpDir
			}

			sim, err := simulator.NewRemoteproc(
				simulator.Config{
					RootDir: rootDir,
					Index:   index,
					Name:    name,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to start simulator: %v", err)
			}
			defer sim.Close()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan
			log.Println("Received shutdown signal")

			return nil
		},
	}

	rootCmd.Flags().UintVar(&index, "index", 0, "is the N in /sys/class/remoteproc/remoteprocN/.../ (default 0)")
	rootCmd.Flags().StringVar(&name, "name", "dsp0", "remote processor name written to /sys/class/remoteproc/.../name")
	rootCmd.Flags().StringVar(&rootDir, "root-dir", "", "location where /sys and /lib will be created")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "show version information")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
