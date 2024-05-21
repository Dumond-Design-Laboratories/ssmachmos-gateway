package server

import (
	"strconv"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/view"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var DATA_SERVICE_UUID = [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD}           // different for every gateway
var DATA_CHARACTERISTIC_UUID = [4]uint32{0x51FF12BB, 0x3ED846E5, 0xB4F9D64E, 0x2FEC021B}    // different for every gateway
var PAIRING_SERVICE_UUID = [4]uint32{0x0000FE59, 0x0000FE59, 0x0000FE59, 0x0000FE59}        // same uuid for every gateway
var PAIRING_CHARACTERISTIC_UUID = [4]uint32{0x0000FE55, 0x0000FE55, 0x0000FE55, 0x0000FE55} // same uuid for every gateway

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
				}},
		}}
	adapter.AddService(&dataService)

	pairing := make(chan bool, 1)
	var pairingCharacteristic bluetooth.Characteristic
	pairingServ := bluetooth.Service{
		UUID: PAIRING_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &pairingCharacteristic,
				UUID:   PAIRING_CHARACTERISTIC_UUID,
				Value:  []byte{}, // the mac address of the ACCEPTED sensor
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					go pairWriteEvent(value, pairingCharacteristic, pairing)
				}}},
	}
	adapter.AddService(&pairingServ)

	adapter.DefaultAdvertisement().Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{
			dataService.UUID,
			pairingServ.UUID,
		}})
}

func StartAdvertising() {
	adapter.DefaultAdvertisement().Start()
	view.Log("Advertising started")
}

func StopAdvertising() {
	adapter.DefaultAdvertisement().Stop()
	view.Log("Advertising stopped")
}

func StartPairing() {
	view.Log("Pairing started")
}

func StopPairing() {
	view.Log("Pairing stopped")
}

func handleWriteData(sensor *model.Sensor, offset int, data []byte) {
	view.Log("Write Event from " + model.MacToString(sensor.Mac))
	view.Log("\tOffset: " + strconv.Itoa(offset))
	view.Log("\tValue: " + string(data))
}

func pairWriteEvent(value []byte, pairingCharacteristic bluetooth.Characteristic, pairing chan bool) {
	if value[0] == 0x00 { // flag => should have one for 1) done pairing 2) pairing request
		// done pairing
		view.Log("Sensor " + string(value) + " has been paired with the Gateway")
		pairingCharacteristic.Write([]byte{})
		<-pairing
	} else if value[0] == 0x01 {
		// pairing request
		view.Log("Sensor " + string(value) + " wants to pair with the Gateway")
		// ask the user if they want to pair
		if true { // condition is temporary until we have a way to ask the user
			pairing <- true // will only allow one sensor to pair at a time (since it's a channel with buffer size 1)
			pairingCharacteristic.Write(value[1:])

			// give 30 second for the sensor to pair, then allow next sensor to pair
			time.Sleep(30 * time.Second)
			if len(pairing) > 0 {
				<-pairing
			}
		}
	}
}
