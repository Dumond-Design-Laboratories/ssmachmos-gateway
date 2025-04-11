# Sensor Suite Machine Monitoring System

The Sensor Suite Machine Monitoring System (SSMachMoS) is a system that manages
MachMoS devices. It consists of a background process that handles device
authentication and configuration, and a GUI to interface with the process.

# Back-end

The back-end is a Linux process written in Go, and depends on a modified fork of
[[https://github.com/tinygo-org/bluetooth][go-bluetooth]]. The server acts as a
BLE GATT peripheral (server) and advertises to potential sensor devices. In this
configuration MachMoS devices scan for devices that advertise the control
service UUID and initiate pairing, as it is less power-intensive than
advertising. The back-end server advertises two services, one for control and
one for data upload. The control service provides a per-sensor configuration
characteristic and a time synchronization characteristic.

## Per-sensor configuration characteristic

The per-sensor configuration characteristic accepts writes from MachMoS devices
and (if not already paired) counts as the start of the pairing process. When a
device configuration is updated, a BLE notification request is sent out to all
devices connected. Devices then respond to the notification and read the
configuration characteristic. A device reading from the per-sensor configuration
characteristic will receive:

- Whether to start sampling, or stay connected on standby
- How long to sleep after each sample and upload cycle
- Sampling configuration for each sensor on-board
  - Should sensor chip be activated
  - Sampling frequency in Hz
  - Sampling duration in seconds
  
## Time synchronization characteristic

The time synchronization characteristic currently simply returns the time the
characteristic read request was received. Later it should involve an NTP-like
procedure for more accurate time measurement.

# Front-end

The front-end is a GUI written in Flutter. It exposes a structured and easy way
to modify the configuration files and a real-time status pane of all known
device. The GUI provides a way to handle device registration without manually
recording the MAC address of MachMoS devices.

# Helpful links for developers

- [[https://docs.arduino.cc/libraries/arduinoble/][Arduino description of a BLE network]]
- [[https://docs.flutter.dev/data-and-backend/state-mgmt/simple][Simple app state management with provider on Flutter]]
- 
