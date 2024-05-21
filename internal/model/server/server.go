package server

import (
	"strconv"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var DATA_SERVICE_UUID = [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD}                 // different for every gateway
var DATA_CHARACTERISTIC_UUID = [4]uint32{0x51FF12BB, 0x3ED846E5, 0xB4F9D64E, 0x2FEC021B}          // different for every gateway
var PAIRING_SERVICE_UUID = [4]uint32{0x0000FE59, 0x0000FE59, 0x0000FE59, 0x0000FE59}              // same uuid for every gateway
var PAIR_REQUEST_CHARACTERISTIC_UUID = [4]uint32{0x0000FE55, 0x0000FE55, 0x0000FE55, 0x0000FE55}  // same uuid for every gateway
var PAIR_RESPONSE_CHARACTERISTIC_UUID = [4]uint32{0x0000FE56, 0x0000FE56, 0x0000FE56, 0x0000FE56} // same uuid for every gateway

type pairingRequest struct {
	mac        [6]byte
	expiration time.Time
}

type PairingState struct {
	active    bool
	requested []pairingRequest
	pairing   [6]byte
}

func Init(sensors *[]model.Sensor) {
	adapter.Enable()

	dataService := bluetooth.Service{
		UUID: DATA_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  DATA_CHARACTERISTIC_UUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					macAddress := [6]byte(value[:6])
					var sensor *model.Sensor
					for _, s := range *sensors {
						if s.Mac == macAddress {
							sensor = &s
							break
						}
					}
					if sensor == nil {
						view.Log("Device " + model.MacToString(macAddress) + " tried to send data but is not authorized")
						return
					}

					go handleWriteData(sensor, offset, value[6:])
				},
			},
		},
	}
	adapter.AddService(&dataService)

	var pairResponse bluetooth.Characteristic
	pairingService := bluetooth.Service{
		UUID: PAIRING_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  PAIR_REQUEST_CHARACTERISTIC_UUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					go pairRequest(value, nil) // TODO pass PairingState
				},
			},
			{
				Handle: &pairResponse,
				UUID:   PAIR_RESPONSE_CHARACTERISTIC_UUID,
				Value:  []byte{}, // the mac address of the ACCEPTED sensor
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					go pairConfirmation(value, pairResponse, nil) // TODO pass PairingState
				},
			},
		},
	}
	adapter.AddService(&pairingService)

	adapter.DefaultAdvertisement().Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{
			dataService.UUID,
			pairingService.UUID,
		},
	})
}

func StartAdvertising() {
	adapter.DefaultAdvertisement().Start()
	view.Log("Advertising started")
}

func StopAdvertising() {
	adapter.DefaultAdvertisement().Stop()
	view.Log("Advertising stopped")
}

func handleWriteData(sensor *model.Sensor, offset int, data []byte) {
	view.Log("Write Event from " + model.MacToString(sensor.Mac))
	view.Log("\tOffset: " + strconv.Itoa(offset))
	view.Log("\tValue: " + string(data))
}

func StartPairing(state *PairingState) {
	state.active = true
	view.Log("Pairing started")
}

func StopPairing(state *PairingState) {
	state.active = false
	view.Log("Pairing stopped")
}

func pairRequest(value []byte, state *PairingState) {
	if len(value) != 6 || !state.active {
		return
	}
	mac := [6]byte(value[:6])
	if state.pairing == mac {
		return
	}
	existed := false
	for _, p := range state.requested {
		if p.mac == mac {
			if time.Now().Before(p.expiration) {
				return
			} else {
				p.expiration = time.Now().Add(30 * time.Second)
				existed = true
				break
			}
		}
	}
	if !existed {
		state.requested = append(state.requested, pairingRequest{
			mac:        mac,
			expiration: time.Now().Add(30 * time.Second),
		})
	}

	view.Log("Pair request from " + model.MacToString(mac) + " | pair <mac-address> to accept")
}

func pairConfirmation(value []byte, pairResponse bluetooth.Characteristic, state *PairingState) {
	if len(value) != 6 || !state.active {
		return
	}
	mac := [6]byte(value[:6])
	if state.pairing != mac {
		return
	}
	state.pairing = [6]byte{}

	pairResponse.Write([]byte{})
	view.Log(model.MacToString(mac) + " has been paired with the Gateway")
}

func Pair(mac [6]byte, pairResponse bluetooth.Characteristic, state *PairingState) {
	if !state.active {
		view.Log("Pairing is not active")
		return
	}
	found := false
	for _, p := range state.requested {
		if p.mac == mac && time.Now().Before(p.expiration) {
			found = true
			break
		}
	}
	if !found {
		view.Log("Pair request from " + model.MacToString(mac) + " not found")
		return
	}
	if state.pairing != [6]byte{} {
		view.Log("Canceled pairing with " + model.MacToString(state.pairing))
		return
	}
	state.pairing = mac

	pairResponse.Write(mac[:])
	view.Log("Pairing with " + model.MacToString(mac))

	go func() {
		time.Sleep(30 * time.Second)
		if state.pairing == mac {
			state.pairing = [6]byte{}
			view.Log("Pairing with " + model.MacToString(mac) + " has timed out")
		}
	}()
}
