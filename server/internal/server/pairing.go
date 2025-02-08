package server

import (
	"encoding/binary"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

type request struct {
	dataTypes          []string
	collectionCapacity uint32
	isPaired           bool
	announcedSensors   bool
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

// Called on device connect
// if device not paired or saved before, starts pair process
// and saves to internal state. Does not get written to disk until BLE agent completes pairing.
// If device is new, returns true
// If device is already paired, returns false
func pairDeviceConnected(MAC [6]byte) bool {
	// Test if MAC address is already stored in settings
	for i := range *Sensors {
		// If device is already written down
		if (*Sensors)[i].Mac == MAC {
			// Fun fact the GUI reads the pairing logs for info. this is awful tbh...
			out.Logger.Println("pairConnectedDevice " + model.MacToString(MAC) + " already exists.")
			return false
		}
	}

	out.Logger.Println("Connected device address " + model.MacToString(MAC))


	// Send out request to GUI
	out.PairingLog("REQUEST-NEW:" + model.MacToString(MAC))

	return true
}

func pairDeviceDisconnected(MAC [6]byte) bool {
	_, ok := state.requested[MAC]
	if ok {
		delete(state.requested, MAC)
		out.PairingLog("PAIR-DEVICE-DISCONNECTED: " + model.MacToString(MAC))
	}

	return true
}

// Device writes out what sensors are on board
func pairReceiveCapabilities(MAC [6]byte, data []byte) bool {
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
	}

	if len(data) != 5 {
		out.Logger.Println("pairReceiveCapabilities expect data with 5 bytes, received", len(data), "bytes instead")
		return false
	}

	// Parse sensor information
	// from protocol.md
	// data types (1 byte) | collection capacity in bytes (4 bytes)
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
	req.announcedSensors = true
	state.requested[MAC] = req

	return true
}

// see protocol.md to understand what is going on here
// internal functions to manage pair lifecycle
// TODO make this triggered by BLE agent on device pair
func pairRequest(value []byte) {
	if len(value) < 12 || !state.active {
		return
	}
	mac := [6]byte(value[1:7])
	for _, s := range *Sensors {
		if s.Mac == mac {
			// Sensor already paired.
			out.PairingLog("REQUEST-SENSOR-EXISTS:" + model.MacToString(mac))
			return
		}
	}
	if _, exists := state.requested[mac]; exists {
		return
	}

	dataTypes := []string{}
	if value[7]&0x01 == 0x01 {
		dataTypes = append(dataTypes, "audio")
	}
	if value[7]&0x02 == 0x02 {
		dataTypes = append(dataTypes, "temperature")
	}
	if value[7]&0x04 == 0x04 {
		dataTypes = append(dataTypes, "vibration")
	}

	collectionCapacity := binary.LittleEndian.Uint32(value[8:12])

	state.requested[mac] = request{
		// publicKey:          publicKey,
		dataTypes:          dataTypes,
		collectionCapacity: collectionCapacity,
	}

	// Announce new sensor pending pairing
	//out.PairingLog("REQUEST-NEW:" + model.MacToString(mac))
}

// see protocol.md to understand what is going on here
// Triggered by sensor to indicate pair done. Remove
func pairConfirmation(mac [6]byte) {
	if _, exists := state.requested[mac]; !exists {
		out.Logger.Println("ERR:Device does not exist")
		return
	}

	// state.pairing = [6]byte{}
	// Display pair code inline with each entry?
	// Or just do it automatically over serial maybe
	// Write sensor data to disk
	model.AddSensor(mac, state.requested[mac].dataTypes, state.requested[mac].collectionCapacity /* state.requested[mac].publicKey, */, Sensors)
	delete(state.requested, mac)

	// I'm 80% sure GUI reads this for pairing information
	out.PairingLog("PAIR-SUCCESS:" + model.MacToString(mac))
}

// see protocol.md to understand what is going on here
// Client requests pairing with this device
// PAIR-ACCEPT $mac
func Pair(mac [6]byte) {
	if !state.active {
		out.PairingLog("PAIRING-DISABLED")
		return
	}
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

	// go func() {
	// 	time.Sleep(30 * time.Second)
	// 	if state.pairing == mac {
	// 		state.pairing = [6]byte{}
	// 		pairResponseCharacteristic.Write([]byte{})
	// 		delete(state.requested, mac)
	// 		out.PairingLog("PAIRING-TIMEOUT:" + model.MacToString(mac))
	// 	}
	// }()
}

func DisconnectDevice(mac [6]byte) {
	//adapter.
}
