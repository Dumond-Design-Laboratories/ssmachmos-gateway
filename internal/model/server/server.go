package server

import (
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var DATA_SERVICE_UUID = [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD}                 // different for every gateway
var DATA_CHARACTERISTIC_UUID = [4]uint32{0x51FF12BB, 0x3ED846E5, 0xB4F9D64E, 0x2FEC021B}          // different for every gateway
var PAIRING_SERVICE_UUID = [4]uint32{0x0000FE59, 0x0000FE59, 0x0000FE59, 0x0000FE59}              // same uuid for every gateway
var PAIR_REQUEST_CHARACTERISTIC_UUID = [4]uint32{0x0000FE55, 0x0000FE55, 0x0000FE55, 0x0000FE55}  // same uuid for every gateway
var PAIR_RESPONSE_CHARACTERISTIC_UUID = [4]uint32{0x0000FE56, 0x0000FE56, 0x0000FE56, 0x0000FE56} // same uuid for every gateway

var DATA_TYPES = map[byte]string{
	0x00: "Vibration",
	0x01: "Acoustic",
	0x02: "Temperature",
	0x03: "Battery",
}

type pairingRequest struct {
	mac        [6]byte
	expiration time.Time
}

type PairingState struct {
	active    bool
	requested []pairingRequest
	pairing   [6]byte
}

func Init(sensors *[]model.Sensor) error {
	err := adapter.Enable()
	if err != nil {
		return err
	}

	dataService := bluetooth.Service{
		UUID: DATA_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  DATA_CHARACTERISTIC_UUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					handleData(client, offset, value, sensors)
				},
			},
		},
	}
	err = adapter.AddService(&dataService)
	if err != nil {
		return err
	}

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
	err = adapter.AddService(&pairingService)
	if err != nil {
		return err
	}

	err = adapter.DefaultAdvertisement().Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{
			dataService.UUID,
			pairingService.UUID,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func StartAdvertising() error {
	err := adapter.DefaultAdvertisement().Start()
	if err != nil {
		return err
	}
	out.Log("Advertising started")
	return nil
}

func StopAdvertising() error {
	err := adapter.DefaultAdvertisement().Stop()
	if err != nil {
		return err
	}
	out.Log("Advertising stopped")
	return nil
}

func handleData(_ bluetooth.Connection, _ int, value []byte, sensors *[]model.Sensor) {
	if len(value) < 8 {
		out.Log("Invalid data format received")
		return
	}

	macAddress := [6]byte(value[:6])
	var sensor *model.Sensor
	for _, s := range *sensors {
		if s.Mac == macAddress {
			sensor = &s
			break
		}
	}
	if sensor == nil {
		out.Log("Device " + model.MacToString(macAddress) + " tried to send data but is not authorized")
		return
	}

	dataType := DATA_TYPES[value[6]]
	samplingFrequency := value[7]

	out.Log("Received data from " + model.MacToString(macAddress) + " (" + sensor.Name + "): " + dataType + " at " + string(samplingFrequency) + "Hz")
}
