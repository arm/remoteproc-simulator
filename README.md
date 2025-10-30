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
./remoteproc-simulator --root-dir /tmp/fake-root --index 0 --name dsp0
```

Control the simulated remote processor via sysfs:

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

Inspect remote processor name:

```bash
cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/name
```

Specify custom firmware load path:

```bash
# Set custom load path
echo /tmp > /tmp/fake-root/sys/module/firmware_class/parameters/path

# Prepare firmware file in custom path
touch /tmp/hi_universe.elf

# Set firmware
echo hi_universe.elf > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware

# Start the remote processor
echo start > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
```

## Installation from Releases

The release binaries are unsigned. On macOS, you'll need to remove the quarantine attribute before running:

```bash
xattr -d com.apple.quarantine ./remoteproc-simulator
```
