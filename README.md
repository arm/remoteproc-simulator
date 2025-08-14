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
./remoteproc-simulator --sysfs /tmp/fake-root/sys --device-index 0 --device-name dsp0
```

Control the simulated device via sysfs:

```bash
# Set firmware
echo 'hello_world.fw' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/firmware

# Start the remote processor
echo 'start' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state

# Check status
cat /tmp/fake-root/sys/class/remoteproc/remoteproc0/state

# Stop the remote processor
echo 'stop' > /tmp/fake-root/sys/class/remoteproc/remoteproc0/state
```
