package server

import (
	"encoding/binary"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

type request struct {
	announcedModel     string
	dataTypes          []string
	collectionCapacity uint32
	isPaired           bool
	announcedSensors   bool
}

type SensorStatus struct {
	Address   string               `json:"address"`
	Name      string               `json:"name"`
	Connected bool                 `json:"connected"`
	LastSeen  time.Time            `json:"last_seen"`
	Activity  model.SensorActivity `json:"activity"`
}

/*
 * List of devices that are connected but aren't fully done yet.
 * Done here means BLE agent pairs and authenticates
 * and device announced sensors on board
 */
type pairingState struct {
	active    bool
	requested map[[6]byte]request
}

// List of all devices pending pairing.
// Here this means devices that are connected, but haven't announced capabilities or finished pairing
var state pairingState

func EnablePairing() {
	state.active = true
	out.Logger.Println("Enabled pairing")
}

func DisablePairing() {
	state.active = false
	out.Logger.Println("Disabled pairing")
}

// Returns list of MAC addresses of devices pending pairing
// A sensor must announce capabilites to show up here
func ListDevicesPendingPairing() []string {
	keys := make([]string, 0, len(state.requested))
	for mac, req := range state.requested {
		if req.announcedSensors == true {
			keys = append(keys, model.MacToString(mac))
		}
	}
	return keys
}

// Returns a list of MAC addresses connected
// GUI compares that to the sensors last seen
func ConnectedDevices() []SensorStatus {
	var devices []SensorStatus = []SensorStatus{}
	// Get all saved sensors
	for _, sensor := range model.Sensors {
		connected := false
		// Intersect with bluetooth connected devices
		for _, dev := range adapter.GetConnectedDevices() {
			if sensor.IsMacEqual(dev.Address.MAC.String()) {
				connected = true
				break
			}
		}
		devices = append(devices, SensorStatus{
			Address:   model.MacToString(sensor.Mac),
			Name:      sensor.Name,
			Connected: connected,
			LastSeen:  sensor.FetchLastSeen().LastSeen,
			Activity:  sensor.FetchLastSeen().LastActivity,
		})
	}
	return devices
}

// If device is already written down
func sensorExists(MAC [6]byte) *model.Sensor {
	for _, sens := range model.Sensors {
		for m := range MAC {
			if MAC[m] != sens.Mac[m] {
				// Get next sensor
				break
			}
			// Exact match, early out with true
			return &sens
		}
	}
	// No sensors matched, false
	return nil
}

// Called on device connect
// if device not paired or saved before, starts pair process
// and saves to internal state. Does not get written to disk until BLE agent completes pairing.
// If device is new, returns true
// If device is already paired, returns false
func pairDeviceConnected(MAC [6]byte) bool {
	// Test if MAC address is already stored in settings
	sensor := sensorExists(MAC)
	if sensor != nil {
		out.Logger.Println("pairDeviceConnected " + model.MacToString(MAC) + " already exists.")
		// Update last seen log
		sensor.UpdateLastSeen(model.SensorActivityIdle)
		// Log that device already exists
		out.Broadcast("SENSOR-CONNECTED:" + model.MacToString(MAC))
		return false
	} else {
		out.Logger.Println("pairConnectedDevice newly discovered " + model.MacToString(MAC))
		// Device is newly connected, give out notification
		//out.Broadcast("PAIR-DEVICE-CONNECTED:" + model.MacToString(MAC))
	}
	out.Logger.Println("Connected device address " + model.MacToString(MAC))

	return true
}

func pairDeviceDisconnected(MAC [6]byte) bool {
	_, ok := state.requested[MAC]
	if ok {
		delete(state.requested, MAC)
		out.Broadcast("PAIR-DEVICE-DISCONNECTED: " + model.MacToString(MAC))
	}
	out.Broadcast("SENSOR-DISCONNECTED:" + model.MacToString(MAC))

	return true
}

// Device writes out what sensors are on board
func pairReceiveCapabilities(MAC [6]byte, data []byte) bool {
	if sensorExists(MAC) != nil {
		return true
	}

	// Go throught with the process only if the sensor is new
	req, ok := state.requested[MAC]

	// If not found
	if !ok {
		// Address is new, store to memory and don't write out until pairing is done + capabilities are out
		// BUG the connect/disconnect callbacks aren't really reliable, might be a BlueZ or library issue
		// In any case, if you write something to this characteristic you're definitely connected
		// and probably know what this bluetooth service does
		state.requested[MAC] = request{
			isPaired:         false, // Depends on user acceptance
			announcedSensors: false, // Depends on sensor giving out details
		}
		req = state.requested[MAC]
		out.Logger.Println("New device", model.MacToString(MAC), "requests pairing")

	}

	if len(data) != 6 {
		out.Logger.Println("pairReceiveCapabilities expect data with 6 bytes, received", len(data), "bytes instead")
		return false
	}

	// Parse sensor information
	// from protocol.md
	// data types (1 byte) | collection capacity in bytes (4 bytes) | sensor model (1 byte)
	// Data types: b(0 0 0 0 0 vibration temperature audio)
	// Save sensors to memory
	bit_types := map[byte]string{
		1 << 0: "audio",
		1 << 1: "temperature",
		1 << 2: "vibration",
		1 << 3: "flux",
	}
	for bit, name := range bit_types {
		// If bit is turned on
		if data[0]&bit == bit {
			req.dataTypes = append(req.dataTypes, name)
		}
	}
	req.collectionCapacity = binary.LittleEndian.Uint32(data[1:5])

	if model, ok := model.SENSOR_MODELS[data[5]]; ok {
		req.announcedModel = model
	} else {
		req.announcedModel = "unknown"
	}

	req.announcedSensors = true
	state.requested[MAC] = req

	// Announce new sensor pending pairing
	out.PairingLog("REQUEST-NEW:" + model.MacToString(MAC))
	out.Logger.Println("Identified sensor", model.MacToString(MAC), ", waiting for pair confirmation")
	return true
}

// see protocol.md to understand what is going on here
// Triggered by sensor to indicate pair done. Remove from pending list
// and notify sensor to collect gateway data
func pairConfirmation(mac [6]byte) {
	if _, exists := state.requested[mac]; !exists {
		out.Logger.Println("ERR:Device does not exist")
		return
	}

	// state.pairing = [6]byte{}
	// Display pair code inline with each entry?
	// Or just do it automatically over serial maybe
	// Write sensor data to disk
	model.AddSensor(mac, state.requested[mac].dataTypes, state.requested[mac].collectionCapacity)
	delete(state.requested, mac)

	// I'm 80% sure GUI reads this for pairing information
	out.Broadcast("PAIR-SUCCESS:" + model.MacToString(mac))
}

// see protocol.md to understand what is going on here
// Client requests pairing with this device
// PAIR-ACCEPT $mac
func Pair(mac [6]byte) {
	if !state.active {
		out.PairingLog("PAIRING-DISABLED")
		return
	}
	out.Logger.Println(state.requested)
	request, ok := state.requested[mac]
	if !ok {
		out.PairingLog("REQUEST-NOT-FOUND " + model.MacToString(mac))
		return
	}
	// TODO start BLE pairing here
	request.isPaired = true
	state.requested[mac] = request
	// For now just mark as paired
	out.PairingLog("PAIRING-WITH:" + model.MacToString(mac))
	out.Logger.Println("Confirming pairing with device")
	pairConfirmation(mac)
}

func DisconnectDevice(mac [6]byte) {
	//adapter.
}
