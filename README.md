# Remoteproc Simulator

A simulator for the Linux remoteproc subsystem that creates a fake sysfs interface for testing purposes. It allows you to simulate remote processor lifecycle management without requiring actual hardware.

## Build

```bash
go build ./cmd/remoteproc-simulator
```

## Test

```bash
go test ./...
```

## Usage

Start the simulator daemon:

```bash
./remoteproc-simulator --root /tmp/fake-root --device-index 0 --device-name dsp0
```

Control the simulated device via sysfs:

```bash
# Prepare firmware file
touch /tmp/fake-root/lib/firmware/hello_world.elf

# Set firmware
echo hello_world.elf > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware

# Start the remote processor
echo start > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state

# Check status
cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state

# Stop the remote processor
echo stop > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
```

Inspect device name:

```bash
cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/name
```

## Installation from Releases

The release binaries are unsigned. On macOS, you'll need to remove the quarantine attribute before running:

```bash
xattr -d com.apple.quarantine ./remoteproc-simulator
```
